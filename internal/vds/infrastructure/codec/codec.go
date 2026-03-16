package codec

import (
	"encoding/json"
	"virturalDevice/internal/vds/domain/message"
)

// Codec 借用json方式来模拟编解码
type Codec struct{}

func NewCodec() *Codec {
	return &Codec{}
}

func (c *Codec) Encode(msg message.Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Codec) Decode(data []byte) (message.Message, error) {
	var msg message.Message
	err := json.Unmarshal(data, &msg)
	return msg, err
}
