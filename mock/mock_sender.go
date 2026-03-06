package mock

import (
	"virturalDevice/connection"
	"virturalDevice/message"
)

// 只发送一次，不管错误的sender
type Sender struct {
}

func (ms *Sender) Send(dst connection.Connection, message message.Message) error {
	err := dst.Send(message.Byte())
	if err != nil {
		return err
	}
	return nil
}
