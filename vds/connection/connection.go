package connection

// Conn 消息通信接口
//
// 注意：具体实现需要允许并发操作. i.e. 多个地方并发调用 Send(), Receive() 要保证不会出错
type Conn interface {
	Send(data []byte) error
	Receive() ([]byte, error)
	Close() error
}
