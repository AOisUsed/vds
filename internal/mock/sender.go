package mock

import (
	"virturalDevice/internal/connection"
)

// 只发送一次，不管错误的sender
type Sender struct {
}

func (ms *Sender) Send(dst connection.Connection, data []byte) error {
	//TODO implement me
	panic("implement me")
}
