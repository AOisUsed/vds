package netconn

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const defaultUDPSocketBufferSize = 256 * 1024

// UDPConnection 通用 UDP 连接封装
// 场景 A (监听者): remoteAddr = nil, 只调用 Receive()
// 场景 B (发送者): remoteAddr != nil, 只调用 Send()
type UDPConnection struct {
	conn       *net.UDPConn
	remoteAddr *net.UDPAddr // 可能为 nil (监听模式)
	localAddr  *net.UDPAddr // 永远不为 nil (创建后由 OS 分配或指定)
	mu         sync.Mutex
	closed     bool
}

// NewUDPConnection 工厂函数
// 根据 Config 自动判断是创建“监听器”还是“发送代理”
func NewUDPConnection(config *Config) (*UDPConnection, error) {
	var remoteAddr *net.UDPAddr
	var localAddr *net.UDPAddr
	var err error

	// 1. 解析远程目标地址 (如果存在)
	if config.Host != "" && config.Port != 0 {
		addrStr := fmt.Sprintf("%s:%d", config.Host, config.Port)
		remoteAddr, err = net.ResolveUDPAddr("udp", addrStr)
		if err != nil {
			return nil, fmt.Errorf("resolve remote address '%s' failed: %w", addrStr, err)
		}
	}

	// 2. 解析本地绑定地址
	lHost := config.LocalHost
	if lHost == "" {
		// 如果没有指定本地 Host，且是发送模式，通常让 OS 自动选择出口网卡 (nil)
		// 如果是监听模式，通常默认监听 0.0.0.0
		if remoteAddr == nil {
			lHost = "0.0.0.0"
		}
	}

	if lHost != "" || config.LocalPort != 0 {
		if lHost == "" {
			lHost = "0.0.0.0"
		}

		ip := net.ParseIP(lHost)
		if ip == nil && lHost != "0.0.0.0" {
			// 尝试解析域名
			ips, err := net.LookupIP(lHost)
			if err != nil || len(ips) == 0 {
				return nil, fmt.Errorf("resolve local host '%s' failed: %w", lHost, err)
			}
			// 优先 IPv4
			for _, i := range ips {
				if i.To4() != nil {
					ip = i.To4()
					break
				}
			}
			if ip == nil {
				ip = ips[0]
			}
		} else if lHost == "0.0.0.0" {
			ip = net.IPv4zero
		}

		localAddr = &net.UDPAddr{
			IP:   ip,
			Port: config.LocalPort, // 0 表示随机
		}
	}

	// 3. 创建 Socket
	// ListenUDP 即使指定了 remoteAddr (第三个参数 nil)，创建的 conn 也是未连接的
	// 它可以接收来自任何地方的数据，也可以发送给任何人 (如果提供了目标地址)
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("listen udp failed: %w", err)
	}

	// 4. 获取实际本地地址 (处理随机端口情况)
	actualLocalAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		_ = conn.Close()
		return nil, errors.New("failed to get local address")
	}

	// 5. 性能优化
	_ = conn.SetReadBuffer(defaultUDPSocketBufferSize)
	_ = conn.SetWriteBuffer(defaultUDPSocketBufferSize)

	return &UDPConnection{
		conn:       conn,
		remoteAddr: remoteAddr, // 如果是监听模式，这里就是 nil
		localAddr:  actualLocalAddr,
		closed:     false,
	}, nil
}

// Send 发送数据
// 如果初始化时指定了 remoteAddr，则发往该地址
// 如果初始化时未指定 (remoteAddr == nil)，则返回错误，因为不知道发给谁
// (注：如果你需要动态发送，可以重载此方法增加 target 参数，但当前设计遵循你的“固定目标”描述)
func (c *UDPConnection) Send(ctx context.Context, data []byte) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return errors.New("connection closed")
	}

	// 关键检查：如果是监听模式创建的连接，没有目标地址，不能直接 Send
	if c.remoteAddr == nil {
		c.mu.Unlock()
		return errors.New("cannot send: no remote address configured (this connection is in listener mode)")
	}

	target := c.remoteAddr
	c.mu.Unlock()

	// 设置超时
	deadline, ok := ctx.Deadline()
	if ok {
		if err := c.conn.SetWriteDeadline(deadline); err != nil {
			return err
		}
	} else {
		if err := c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
			return err
		}
	}

	_, err := c.conn.WriteToUDP(data, target)
	return err
}

// Receive 接收数据
// 适用于监听模式
// 如果在发送模式 (有固定 remoteAddr) 下调用，理论上也能收到包 (UDP socket 特性)，
// 但为了逻辑清晰，可以选择性地在 remoteAddr != nil 时警告或阻止，或者放任不管。
// 这里选择放任不管，因为 UDP socket 确实可以既发又收，只是业务逻辑上不建议混用。
func (c *UDPConnection) Receive(ctx context.Context) ([]byte, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, errors.New("connection closed")
	}
	c.mu.Unlock()

	// 可选：如果这是纯发送代理 (remoteAddr != nil)，可以选择不让它 Receive
	// 但为了通用性，这里不强制报错，除非你有严格的安全需求
	// if c.remoteAddr != nil { return nil, errors.New("this connection is intended for sending only") }

	deadline, ok := ctx.Deadline()
	if ok {
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			return nil, err
		}
	} else {
		// 默认长超时，避免永久阻塞
		if err := c.conn.SetReadDeadline(time.Now().Add(1 * time.Hour)); err != nil {
			return nil, err
		}
	}

	buf := make([]byte, 4096)
	n, _, err := c.conn.ReadFromUDP(buf)
	if err != nil {
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

// Close 关闭连接
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
// 目的是用于把自身的UDP地址Config的形式保存，存在数据库中，以供其他设备连接。
//
// 因此Config中的Host是自己本地地址，Port是自己本地端口。
func (c *UDPConnection) Config() *Config {
	cfg := &Config{
		Type: "udp",
		Host: c.localAddr.IP.String(),
		Port: c.localAddr.Port,
	}
	return cfg
}
