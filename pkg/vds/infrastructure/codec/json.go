package codec

import (
	"encoding/json"
	"virturalDevice/pkg/vds/domain/message"
)

// JsonCodec 借用json方式来模拟编解码
type JsonCodec struct{}

func NewCodec() *JsonCodec {
	return &JsonCodec{}
}

func (c *JsonCodec) Encode(msg message.Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *JsonCodec) Decode(data []byte) (message.Message, error) {
	var msg message.Message
	err := json.Unmarshal(data, &msg)
	return msg, err
}
