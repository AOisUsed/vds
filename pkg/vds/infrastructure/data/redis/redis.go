package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config 持有 Redis 配置信息
type Config struct {
	Addr     string
	Password string
	DB       int
}

// Manager 封装了 Redis 客户端及其生命周期操作方法
// 这是一个纯 Go 结构体，不依赖 fx 框架即可使用
type Manager struct {
	Client *redis.Client
	Config Config
}

// NewManager 创建管理器实例（但不立即连接）
// 这是“手动控制”模式的入口
func NewManager(cfg Config) *Manager {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &Manager{
		Client: client,
		Config: cfg,
	}
}

// Connect 手动执行连接和健康检查
// 对应 fx 中的 OnStart 逻辑
func (m *Manager) Connect(ctx context.Context) error {
	fmt.Println("🔌  正在连接 Redis...")

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := m.Client.Ping(pingCtx).Result()
	if err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}

	fmt.Println("✅  Redis 连接成功!")
	return nil
}

// Close 手动优雅关闭连接
// 对应 fx 中的 OnStop 逻辑
func (m *Manager) Close(ctx context.Context) error {
	fmt.Println("🛑  正在关闭 Redis 连接...")

	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := m.Client.Close(); err != nil {
		return fmt.Errorf("redis close error: %w", err)
	}

	fmt.Println("✅  Redis 连接已关闭")
	return nil
}

// GetClient 暴露底层的 redis.Client 供业务逻辑使用
func (m *Manager) GetClient() *redis.Client {
	return m.Client
}
