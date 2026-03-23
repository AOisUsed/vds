package connection

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const defaultUDPSocketBufferSize = 256 * 1024

// UDPConnection 实现了 Connection 接口, Configurable 接口
type UDPConnection struct {
	conn   *net.UDPConn
	addr   *net.UDPAddr
	mu     sync.Mutex // 保护并发读写操作，虽然 net.UDPConn 部分线程安全，但逻辑层串行化更安全
	closed bool
}

// NewUDPConnection 创建一个新的 UDP 连接
// 根据 config 中的信息初始化，并应用默认的性能优化参数
func NewUDPConnection(config *Config) (*UDPConnection, error) {
	// 1. 解析远程地址 (支持 IP 或 域名)
	remoteAddrStr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	remoteAddr, err := net.ResolveUDPAddr("udp", remoteAddrStr)
	if err != nil {
		return nil, fmt.Errorf("resolve remote address '%s' failed: %w", remoteAddrStr, err)
	}

	// 2. 准备本地地址 (仅在配置了 LocalPort 或 LocalHost 时生效)
	var localAddr *net.UDPAddr
	if config.LocalPort != 0 || config.LocalHost != "" {
		lHost := config.LocalHost
		if lHost == "" {
			lHost = "0.0.0.0"
		}

		// 注意：LocalHost 通常期望是 IP。如果是域名，这里也需要 Resolve。
		// 为了简单起见，这里假设 LocalHost 是 IP 字符串。如果是 "0.0.0.0"，ParseIP 能正确处理。
		ip := net.ParseIP(lHost)
		if ip == nil && lHost != "0.0.0.0" {
			// 如果用户填了域名作为本地绑定地址（较少见），尝试解析
			// 大多数情况下本地绑定直接用 IP 或 0.0.0.0
			ips, err := net.LookupIP(lHost)
			if err != nil || len(ips) == 0 {
				return nil, fmt.Errorf("resolve local host '%s' failed: %w", lHost, err)
			}
			ip = ips[0]
		}

		localAddr = &net.UDPAddr{
			IP:   ip,
			Port: config.LocalPort,
		}
	}

	// 3. 监听/创建 Socket
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("listen udp failed (local: %v): %w", localAddr, err)
	}

	// 4. 应用性能优化 (硬编码在内部，方便统一调整)

	// 设置读缓冲区
	if err = conn.SetReadBuffer(defaultUDPSocketBufferSize); err != nil {
		// fmt.Printf("Warning: failed to set read buffer to %d: %v\n", defaultUDPSocketBufferSize, err)
	}

	// 设置写缓冲区
	if err = conn.SetWriteBuffer(defaultUDPSocketBufferSize); err != nil {
		// 同上
	}

	return &UDPConnection{
		conn:   conn,
		addr:   remoteAddr,
		closed: false,
	}, nil
}

// Send 实现 Connection 接口
func (c *UDPConnection) Send(ctx context.Context, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("connection closed")
	}

	// 设置写超时，防止阻塞
	deadline, ok := ctx.Deadline()
	if ok {
		if err := c.conn.SetWriteDeadline(deadline); err != nil {
			return err
		}
	} else {
		// 默认超时 5 秒
		if err := c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
			return err
		}
	}

	_, err := c.conn.WriteToUDP(data, c.addr)
	return err
}

// Receive 实现 Connection 接口
func (c *UDPConnection) Receive(ctx context.Context) ([]byte, error) {
	c.mu.Lock()
	// 注意：不要在持有锁的时候无限期阻塞等待网络 IO，否则 Close() 会被阻塞
	// 但 net.UDPConn 的 Read 在另一个 goroutine 调用 Close 时会返回错误，所以这里可以持锁
	// 为了更稳健，我们只在设置读死限时持锁，或者利用 context 控制

	// 实际上，net.Conn 的 Read/Write 在并发调用 Close 时是安全的，会立即返回错误。
	// 所以这里的锁主要是为了防止业务逻辑层面的竞态，而非底层 IO。
	c.mu.Unlock()

	if c.closed {
		return nil, fmt.Errorf("connection closed")
	}

	// 设置读超时以响应 Context
	deadline, ok := ctx.Deadline()
	if ok {
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			return nil, err
		}
	} else {
		// 如果没有上下文超时，给一个较长的默认超时，或者由调用方控制
		// 这里设为 1 小时，避免永久阻塞，实际应由 ctx 控制
		if err := c.conn.SetReadDeadline(time.Now().Add(1 * time.Hour)); err != nil {
			return nil, err
		}
	}

	buf := make([]byte, 4096)
	n, _, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		// 如果是超时错误，且上下文已取消，返回上下文错误
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}
		return nil, err
	}

	return buf[:n], nil
}

// Close 实现 Connection 接口
func (c *UDPConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.conn.Close()
}

// Config 实现 Configurable 接口
func (c *UDPConnection) Config() *Config {
	return &Config{
		Host: c.addr.String(),
		Port: c.addr.Port,
		Type: c.addr.Network(),
	}
}
