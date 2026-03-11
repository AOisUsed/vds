package mock

import (
	"virturalDevice/internal/connection"
)

// 只发送一次，不管错误的sender
type Sender struct{}

func NewSender() *Sender {
	return new(Sender)
}

func (ms *Sender) Send(dst connection.Connection, data []byte) error {
	err := dst.Send(data)
	return err
}
