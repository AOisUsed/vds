package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"virturalDevice/pkg/vds/domain/connection"
	"virturalDevice/pkg/vds/domain/repository"
	"virturalDevice/pkg/vds/domain/virtualdevice/params"
	"virturalDevice/pkg/vds/infrastructure/connection/netconn"
	"virturalDevice/pkg/vds/infrastructure/deviceparams"

	"github.com/redis/go-redis/v9"
)

// 定义 Redis Key 的前缀，方便管理
const (
	keyPrefixParams     = "vd:params:"
	keyPrefixConnection = "vd:conn:"
)

// RedisVDRepo 使用redis作为底层数据库的数据仓库，
// 使用 connStore 来把redis中的连接配置转化为真实的连接输出，
// 以及把真实的连接类型转化为配置持久化到redis中。
type RedisVDRepo struct {
	client    *redis.Client
	connStore *netconn.Store
}

// NewRedisVDRepo 构造函数
// 需要传入一个已经初始化好的 Redis Client
func NewRedisVDRepo(client *redis.Client, store *netconn.Store) repository.VDRepository {
	return &RedisVDRepo{
		client:    client,
		connStore: store,
	}
}

// --- 辅助方法：序列化与反序列化 ---

func marshal(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshal(data string, v any) error {
	if data == "" {
		return fmt.Errorf("data is empty")
	}
	return json.Unmarshal([]byte(data), v)
}

// --- 实现 Params 相关方法 ---

func (r *RedisVDRepo) Params(ctx context.Context, id string) (params.Params, error) {

	key := fmt.Sprintf("%s%s", keyPrefixParams, id)

	// 获取 params 值
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("params not found for id: %s", id)
	}
	if err != nil {
		return nil, err
	}

	var t struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal([]byte(val), &t)

	var p params.Params

	switch t.Type {
	case "RadioParams":
		p = &deviceparams.RadioParams{}
	// case "..." : .....   // 如果后续需要扩展，可以添加其他种类Params,
	default:
		p = params.NewEmpty()
	}

	if err = unmarshal(val, &p); err != nil {
		return p, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	return p, nil
}

func (r *RedisVDRepo) SetParams(ctx context.Context, id string, p params.Params) error {
	key := fmt.Sprintf("%s%s", keyPrefixParams, id)

	val, err := marshal(p)
	if err != nil {
		return err
	}

	// 这里可以选择设置过期时间，例如 24 小时，防止脏数据永久留存
	return r.client.Set(ctx, key, val, time.Hour).Err()
}

func (r *RedisVDRepo) RemoveParams(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s%s", keyPrefixParams, id)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisVDRepo) AllParams(ctx context.Context) (map[string]params.Params, error) {
	result := make(map[string]params.Params)
	pattern := fmt.Sprintf("%s*", keyPrefixParams)

	cursor := uint64(0)
	const batchSize = 100

	// 使用 Scan 遍历所有匹配的 key，避免 Keys 命令阻塞生产环境
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, batchSize).Result()
		if err != nil {
			return nil, err
		}

		if len(keys) > 0 {
			// 批量获取值 (Pipeline 优化性能)
			pipe := r.client.Pipeline()
			cmds := make([]*redis.StringCmd, len(keys))

			for i, key := range keys {
				cmds[i] = pipe.Get(ctx, key)
			}

			_, err = pipe.Exec(ctx)
			if err != nil && !errors.Is(err, redis.Nil) {
				// 部分失败处理视业务需求而定，这里简单返回错误
				// 也可以继续处理成功的部分
			}

			for i, key := range keys {
				val, err := cmds[i].Result()
				if errors.Is(err, redis.Nil) {
					continue // 键在获取瞬间被删除了
				}
				if err != nil {
					// 记录日志或跳过
					continue
				}

				var p params.Params
				if err := unmarshal(val, &p); err != nil {
					// 数据损坏，跳过或记录日志
					continue
				}

				// 提取 ID: key 是 "vd:params:123", 我们需要 "123"
				id := key[len(keyPrefixParams):]
				result[id] = p
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return result, nil
}

// --- 实现 Connection 相关方法 ---

func (r *RedisVDRepo) Connection(ctx context.Context, id string) (connection.Connection, error) {

	// 从数据库获取连接的配置文件
	key := fmt.Sprintf("%s%s", keyPrefixConnection, id)
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("没有找到设备id %s 的连接", id)
	}
	if err != nil {
		return nil, err
	}

	var config netconn.Config
	if err = unmarshal(val, &config); err != nil {
		return nil, fmt.Errorf("无法反序列化连接配置: %w", err)
	}

	// 使用 conn store 获取连接
	conn, err := r.connStore.GetConnection(&config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (r *RedisVDRepo) SetConnection(ctx context.Context, id string, conn connection.Connection) error {
	key := fmt.Sprintf("%s%s", keyPrefixConnection, id)

	configurableConn, ok := conn.(netconn.Configurable)
	if !ok {
		return fmt.Errorf("连接方式没有实现configurable接口，无法持久化到数据库中")
	}

	// 把连接的配置文件，而不是其本身持久化到redis中
	val, err := marshal(configurableConn.Config())
	if err != nil {
		return err
	}

	// 存入redis
	// 连接信息通常有过期时间，比如心跳超时时间 * 2
	// 这里暂设 1 小时，具体看业务需求
	return r.client.Set(ctx, key, val, time.Hour).Err()
}

func (r *RedisVDRepo) RemoveConnection(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s%s", keyPrefixConnection, id)
	return r.client.Del(ctx, key).Err()
}
