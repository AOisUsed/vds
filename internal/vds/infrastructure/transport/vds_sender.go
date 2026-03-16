package transport

import (
	"virturalDevice/internal/vds/domain/connection"
)

// Sender 只发送一次，不进行任何处理的sender
type Sender struct{}

func NewSender() *Sender {
	return new(Sender)
}

func (ms *Sender) Send(dst connection.Connection, data []byte) error {
	err := dst.Send(data)
	return err
}
