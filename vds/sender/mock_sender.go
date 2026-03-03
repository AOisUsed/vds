package sender

import (
	"virturalDevice/message"
	"virturalDevice/vds/address"
)

// 只发送一次，不管错误的sender
type mockSender struct {
}

func (ms *mockSender) Send(dst address.Address, message message.Message) error {
	err := dst.Send(message.Byte())
	if err != nil {
		return err
	}
	return nil
}
