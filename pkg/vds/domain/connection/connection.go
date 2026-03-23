package connection

import "context"

// Connection 消息通信接口
//
// 注意：具体实现需要允许并发操作. i.e. 多个地方并发调用 Send() / Receive() 要保证不会出错
type Connection interface {
	// Send 发送数据。实现应处理上下文取消。
	Send(ctx context.Context, data []byte) error

	// Receive 接收数据。实现应阻塞直到收到数据或上下文取消。
	Receive(ctx context.Context) ([]byte, error)

	// Close 关闭底层连接。
	Close() error
}
