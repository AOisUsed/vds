package sender

import (
	"virturalDevice/message"
	"virturalDevice/vds/address"
)

type Sender interface {
	Send(dst address.Address, message message.Message) error
}
