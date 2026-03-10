package codec

import "virturalDevice/internal/message"

type Codec interface {
	Encode(msg message.Message) ([]byte, error)
	Decode(data []byte) (message.Message, error)
}
