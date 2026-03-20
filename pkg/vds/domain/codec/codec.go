package codec

import (
	"virturalDevice/pkg/vds/domain/message"
)

type Codec interface {
	Encode(msg message.Message) ([]byte, error)
	Decode(data []byte) (message.Message, error)
}
