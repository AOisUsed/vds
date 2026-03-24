package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// Module 导出一个标准的 fx.Module
// 引入此模块即可自动获得 *redis.Client 的依赖注入和生命周期管理
var Module = fx.Module("redis",
	// 1. 提供配置 (实际项目中可从 viper 注入)
	fx.Provide(func() Config {
		return Config{
			Addr:     "192.168.0.175:6379",
			Password: "",
			DB:       15,
		}
	}),

	// 2. 提供 RedisManager (包装了连接逻辑的核心对象)
	// 注意：这里只提供管理器，业务代码通常直接依赖 *redis.Client
	fx.Provide(NewManager),

	// 3. 提供 *redis.Client 给其他组件使用
	// 依赖 RedisManager，通过管理器获取客户端
	fx.Provide(func(m *Manager) *redis.Client {
		return m.GetClient()
	}),

	// 4. 注册生命周期钩子
	// 当应用启动时，自动调用 manager.Connect
	// 当应用停止时，自动调用 manager.Close
	fx.Invoke(func(lc fx.Lifecycle, m *Manager) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return m.Connect(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return m.Close(ctx)
			},
		})
	}),
)
