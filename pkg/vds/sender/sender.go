package sender

import (
	"virturalDevice/pkg/connection"
	"virturalDevice/pkg/message"
)

// Sender vds内统一消息发送器
type Sender interface {
	Send(dst connection.Connection, message message.Message) error
}
