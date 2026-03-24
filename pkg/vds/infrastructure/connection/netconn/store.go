package netconn

import (
	"fmt"
	"sync"
	"virturalDevice/pkg/vds/domain/connection"
)

// Store 管理所有活动连接的生命周期
// 使用 map[string]Connection 存储，Key 为标准化的 "IP:Port" 字符串
type Store struct {
	connectionsByAddr map[string]connection.Connection
	mu                sync.RWMutex // 读写锁：读多写少场景性能更好
}

// NewStore 初始化连接商店
func NewStore() *Store {
	return &Store{
		connectionsByAddr: make(map[string]connection.Connection),
	}
}

// CreateConnection 创建并注册一个新连接
// 如果该地址已存在连接，则返回错误（避免重复创建）
func (s *Store) CreateConnection(config *Config) (connection.Connection, error) {

	var addrStr string
	var conn connection.Connection

	// 根据类型工厂模式创建 (保留扩展能力)
	switch config.Type {
	case "udp":
		addrStr = fmt.Sprintf("%s:%d", config.Host, config.Port)

		c, err := NewUDPConnection(config)
		if err != nil {
			return nil, fmt.Errorf("创建udp连接失败: %w", err)
		}
		conn = c
	default:
		return nil, fmt.Errorf("不支持的连接方式: %s", config.Type)
	}

	// 存入 Map
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.connectionsByAddr[addrStr]; exists {
		return nil, fmt.Errorf("和 %s 的连接已经存在", addrStr)
	}

	s.connectionsByAddr[addrStr] = conn
	return conn, nil
}

// GetConnection 通过地址获取连接,如果不存在则创建并返回创建的连接
func (s *Store) GetConnection(config *Config) (connection.Connection, error) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	s.mu.RLock()
	conn, exists := s.connectionsByAddr[addr]
	s.mu.RUnlock()

	if exists {
		return conn, nil
	}

	// 如果不存在连接的情况,创建连接
	conn, err := s.CreateConnection(config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// CloseAndRemove 关闭指定地址的连接并从 Store 中移除
// 这是生命周期管理的核心：先关闭资源，再删除引用
func (s *Store) CloseAndRemove(config *Config) error {
	var err error

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.connectionsByAddr[addr]
	if !exists {
		return fmt.Errorf("找不到和 %s 的连接, 无法关闭", addr)
	}

	if err = conn.Close(); err != nil {
		// 即使关闭出错，也要从 Map 中移除
		fmt.Printf("警告: 无法关闭连接 %s: %v\n", addr, err)
	}

	// 从 Map 移除
	delete(s.connectionsByAddr, addr)
	return err
}

// CloseAll 关闭所有连接并清空 Store
// 常用于程序退出时的清理工作
func (s *Store) CloseAll() error {
	s.mu.Lock()
	// 提取所有连接到一个临时切片，释放锁后再逐个关闭
	// 这样可以避免长时间持有锁，同时允许 Close 操作并发执行（如果底层支持）
	connsToClose := make([]connection.Connection, 0, len(s.connectionsByAddr))
	for _, conn := range s.connectionsByAddr {
		connsToClose = append(connsToClose, conn)
	}

	// 清空 Map
	s.connectionsByAddr = make(map[string]connection.Connection)
	s.mu.Unlock()

	// 并发或串行关闭均可，这里串行处理以简化错误收集
	var firstErr error
	for _, conn := range connsToClose {
		if err := conn.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			// 可以记录其他错误到日志
		}
	}

	return firstErr
}

// Count 返回当前活跃连接数量
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connectionsByAddr)
}

// ListAddresses 列出所有已连接的地址字符串 (用于调试或监控)
func (s *Store) ListAddresses() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	addrs := make([]string, 0, len(s.connectionsByAddr))
	for addr := range s.connectionsByAddr {
		addrs = append(addrs, addr)
	}
	return addrs
}
