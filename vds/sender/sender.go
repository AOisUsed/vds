package sender

import (
	"virturalDevice/message"
	"virturalDevice/vds/connection"
)

// Sender vds内统一消息发送器
type Sender interface {
	Send(dst connection.Conn, message message.Message) error
}
