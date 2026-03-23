package sender

import (
	"context"
	"virturalDevice/pkg/vds/domain/connection"
)

// Sender 只发送一次，不进行任何处理的sender
type Sender struct{}

func NewSender() *Sender {
	return new(Sender)
}

func (ms *Sender) Send(ctx context.Context, conn connection.Connection, data []byte) error {
	err := conn.Send(ctx, data)
	return err
}
