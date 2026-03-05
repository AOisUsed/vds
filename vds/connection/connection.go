package connection

// Conn 消息通信接口
type Conn interface {
	Send(data []byte) error
	Receive() ([]byte, error)
	Close() error
}
