package connection

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// UDPConfig 作为建立UDP连接的参数，也可以
type UDPConfig struct {
	LocalIP       string `mapstructure:"local_ip"`    // "0.0.0.0"
	LocalPort     int    `mapstructure:"local_port"`  // 8080
	RemoteIP      string `mapstructure:"remote_ip"`   // 可选，如果为空则不绑定远程
	RemotePort    int    `mapstructure:"remote_port"` // 可选
	BufferSize    int    `mapstructure:"buffer_size"` // 建议 65535
	ReadTimeoutMs int    `mapstructure:"read_timeout_ms"`
}

// UDPConn 包装了 net.UDPConn
type UDPConn struct {
	*net.UDPConn
	config     *UDPConfig
	remoteAddr *net.UDPAddr // 如果配置了远程地址，这里不为 nil
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	isClosed   bool
	mu         sync.Mutex
}

// NewUDPConn 工厂函数
// 注意：这里只做初始化，不启动接收循环。启动逻辑在 Serve 中。
func NewUDPConn(cfg *UDPConfig) (*UDPConn, error) {
	// 1. 解析本地地址
	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", cfg.LocalIP, cfg.LocalPort))
	if err != nil {
		return nil, fmt.Errorf("解析本地地址失败: %w", err)
	}
	var conn *net.UDPConn
	var remoteAddr *net.UDPAddr

	// 2. 判断是否需要关联远程地址
	if cfg.RemoteIP != "" && cfg.RemotePort > 0 {
		// --- 场景 A: 需要关联远程地址 (类似 TCP 的行为) ---
		remoteAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", cfg.RemoteIP, cfg.RemotePort))
		if err != nil {
			return nil, fmt.Errorf("解析远程地址失败: %w", err)
		}

		c, err := net.DialUDP("udp", localAddr, remoteAddr)
		if err != nil {
			return nil, fmt.Errorf("dial UDP 失败: %w", err)
		}
		conn = c

		// 此时 conn 已经“连接”到了 remoteAddr
		// Write(data) 会自动发往 remoteAddr
		// Read() 只会收到来自 remoteAddr 的数据

	} else {
		// --- 场景 B: 纯服务端模式 (接收所有人的数据) ---
		c, err := net.ListenUDP("udp", localAddr)
		if err != nil {
			return nil, fmt.Errorf("listen UDP 失败: %w", err)
		}
		conn = c
		// 此时 remoteAddr 为 nil
		// 必须使用 WriteToUDP(data, addr) 发送
		// ReadFromUDP() 会返回发送者的地址
	}

	// 3. 设置读缓冲区 (非常重要，防止丢包)
	if cfg.BufferSize > 0 {
		if err := conn.SetReadBuffer(cfg.BufferSize); err != nil {
			fmt.Printf("警告：设置读缓冲区失败: %v\n", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &UDPConn{
		UDPConn:    conn,
		config:     cfg,
		remoteAddr: remoteAddr,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Serve 启动接收循环 (对应 Fx 的 OnStart)
// 这通常是一个阻塞操作，所以要在 goroutine 中运行，或者由业务层决定如何处理
func (u *UDPConn) Serve() error {
	u.wg.Add(1)
	go func() {
		defer u.wg.Done()

		buffer := make([]byte, u.config.BufferSize)
		if buffer == nil || len(buffer) == 0 {
			buffer = make([]byte, 65535) // 默认 fallback
		}

		timeout := time.Duration(u.config.ReadTimeoutMs) * time.Millisecond
		if timeout == 0 {
			timeout = 5 * time.Second // 默认超时
		}

		for {
			// 设置读超时，以便能定期检查 ctx 是否被取消
			u.SetReadDeadline(time.Now().Add(timeout))

			n, addr, err := u.ReadFromUDP(buffer)

			// 检查是否是关闭信号或超时
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					// 超时是正常的，检查是否需要退出
					select {
					case <-u.ctx.Done():
						fmt.Println("UDP 接收循环收到退出信号")
						return
					default:
						continue // 继续下一次读取
					}
				}

				// 其他错误（如连接已关闭）
				if u.isClosed {
					return
				}
				fmt.Printf("UDP 读取错误: %v\n", err)
				continue
			}

			// 处理收到的数据 (这里只是打印，实际应该发送到 channel 或回调)
			data := buffer[:n]
			// TODO: 将 data 和 addr 传递给业务逻辑处理
			// u.handleData(data, addr)
			_ = data
			_ = addr
		}
	}()

	return nil
}

// Send 实现接口
func (u *UDPConn) Send(data []byte) error {
	if u.remoteAddr != nil {
		// 如果已经 Connect 过，直接用 Write
		_, err := u.Write(data)
		return err
	}

	// 如果没有绑定远程地址，Send 方法无法知道发给谁
	// 这种情况设计上有问题，除非 Send 方法签名改为 Send(addr, data)
	return fmt.Errorf("未配置远程地址，无法直接 Send")
}

// Receive 实现接口 (注意：UDP 的 Receive 通常是阻塞的，或者通过 Channel 异步获取)
// 这里的实现比较 tricky，因为 UDP 是消息驱动的。
// 更好的模式是 Serve() 内部处理接收，然后通过 Channel 推送给消费者。
// 如果必须实现这个接口，可以做一个简单的阻塞读（不推荐用于高并发）
func (u *UDPConn) Receive() ([]byte, error) {
	buffer := make([]byte, u.config.BufferSize)
	u.SetReadDeadline(time.Time{}) // 清除超时，永久阻塞直到收到数据或关闭
	n, _, err := u.ReadFromUDP(buffer)
	if err != nil {
		return nil, err
	}
	return buffer[:n], nil
}

// Close 实现接口 (对应 Fx 的 OnStop)
func (u *UDPConn) Close() error {
	u.mu.Lock()
	if u.isClosed {
		u.mu.Unlock()
		return nil
	}
	u.isClosed = true
	u.mu.Unlock()

	// 1. 取消上下文，通知 Serve 循环退出
	u.cancel()

	// 2. 关闭底层连接 (这会立即导致 ReadFromUDP 返回错误，唤醒阻塞的读)
	if err := u.UDPConn.Close(); err != nil {
		return err
	}

	// 3. 等待后台 Goroutine 结束
	u.wg.Wait()

	fmt.Println("UDP 连接已完全关闭")
	return nil
}
