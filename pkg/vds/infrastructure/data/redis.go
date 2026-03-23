package data

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// Config 持有 Redis 配置信息
// 在实际项目中，这通常从配置文件或环境变量注入
type Config struct {
	Addr     string
	Password string
	DB       int
}

// NewConfig 提供一个默认配置，实际使用中可以从 viper 等配置中心注入
func NewConfig() Config {
	return Config{
		Addr:     "192.168.0.175:6379",
		Password: "", // 如果有密码请设置
		DB:       15,
	}
}

// ClientParams 是传递给构造函数的参数结构体
// 使用 fx.In 可以方便地接收依赖
type ClientParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Config    Config
}

// NewClient 是 Redis 客户端的构造函数
// 它会被 fx 容器调用，并将生成的 *redis.Client 提供给其他组件
func NewClient(params ClientParams) *redis.Client {
	// 1. 创建客户端实例 (此时尚未连接)
	client := redis.NewClient(&redis.Options{
		Addr:     params.Config.Addr,
		Password: params.Config.Password,
		DB:       params.Config.DB,
	})

	// 2. 注册生命周期钩子
	// OnStart: 应用启动时执行，用于健康检查或预热连接
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("🔌 正在连接 Redis...")
			// 创建一个带超时的上下文进行 Ping 测试
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := client.Ping(pingCtx).Result()
			if err != nil {
				return fmt.Errorf("redis connection failed: %w", err)
			}

			fmt.Println("✅ Redis 连接成功!")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("🛑 正在关闭 Redis 连接...")
			// 优雅关闭连接
			_, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err := client.Close(); err != nil {
				return fmt.Errorf("redis close error: %w", err)
			}
			fmt.Println("✅ Redis 连接已关闭")
			return nil
		},
	})

	// 3. 返回客户端实例
	// 注意：这里返回的是客户端指针，其他组件可以通过依赖注入获取它
	return client
}
