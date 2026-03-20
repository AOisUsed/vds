package sender

import (
	"virturalDevice/pkg/vds/domain/connection"
)

// Sender vds内统一消息发送器
type Sender interface {
	Send(dst connection.Connection, data []byte) error
}
