package mock

import (
	"virturalDevice/message"
	"virturalDevice/vds/connection"
)

// 只发送一次，不管错误的sender
type mockSender struct {
}

func (ms *mockSender) Send(dst connection.Conn, message message.Message) error {
	err := dst.Send(message.Byte())
	if err != nil {
		return err
	}
	return nil
}
