package sender

import (
	"virturalDevice/internal/vds/domain/connection"
)

// Sender vds内统一消息发送器
type Sender interface {
	Send(dst connection.Connection, data []byte) error
}
