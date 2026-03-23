package sender

import (
	"context"
	"virturalDevice/pkg/vds/domain/connection"
)

// Sender vds内统一消息发送器
type Sender interface {
	Send(ctx context.Context, conn connection.Connection, data []byte) error
}
