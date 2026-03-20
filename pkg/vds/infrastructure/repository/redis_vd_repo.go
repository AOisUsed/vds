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

	"github.com/redis/go-redis/v9"
)

// 定义 Redis Key 的前缀，方便管理
const (
	keyPrefixParams     = "vd:params:"
	keyPrefixConnection = "vd:conn:"
)

// redisVDRepo 实现
type redisVDRepo struct {
	client *redis.Client
}

// NewRedisVDRepo 构造函数
// 需要传入一个已经初始化好的 Redis Client
func NewRedisVDRepo(client *redis.Client) repository.VDRepository {
	return &redisVDRepo{
		client: client,
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

func (r *redisVDRepo) Params(ctx context.Context, id string) (params.Params, error) {
	var p params.Params
	key := fmt.Sprintf("%s%s", keyPrefixParams, id)

	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return p, fmt.Errorf("params not found for id: %s", id)
	}
	if err != nil {
		return p, err
	}

	if err := unmarshal(val, &p); err != nil {
		return p, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	return p, nil
}

func (r *redisVDRepo) SetParams(ctx context.Context, id string, p params.Params) error {
	key := fmt.Sprintf("%s%s", keyPrefixParams, id)

	val, err := marshal(p)
	if err != nil {
		return err
	}

	// 这里可以选择设置过期时间，例如 24 小时，防止脏数据永久留存
	return r.client.Set(ctx, key, val, time.Hour).Err()
}

func (r *redisVDRepo) RemoveParams(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s%s", keyPrefixParams, id)
	return r.client.Del(ctx, key).Err()
}

func (r *redisVDRepo) AllParams(ctx context.Context) (map[string]params.Params, error) {
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
// 逻辑与 Params 完全一致，只是 Key 前缀和类型不同

func (r *redisVDRepo) Connection(ctx context.Context, id string) (connection.Connection, error) {
	var c connection.Connection
	key := fmt.Sprintf("%s%s", keyPrefixConnection, id)

	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return c, fmt.Errorf("connection not found for id: %s", id)
	}
	if err != nil {
		return c, err
	}

	if err := unmarshal(val, &c); err != nil {
		return c, fmt.Errorf("failed to unmarshal connection: %w", err)
	}

	return c, nil
}

func (r *redisVDRepo) SetConnection(ctx context.Context, id string, conn connection.Connection) error {
	key := fmt.Sprintf("%s%s", keyPrefixConnection, id)

	val, err := marshal(conn)
	if err != nil {
		return err
	}

	// 连接信息通常有过期时间，比如心跳超时时间 * 2
	// 这里暂设 1 小时，具体看业务需求
	return r.client.Set(ctx, key, val, time.Hour).Err()
}

func (r *redisVDRepo) RemoveConnection(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s%s", keyPrefixConnection, id)
	return r.client.Del(ctx, key).Err()
}
