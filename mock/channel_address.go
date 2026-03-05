package mock

import (
	"errors"
	"time"
)

type chanAddr struct {
	messageCh chan []byte
}

func (c chanAddr) Send(data []byte) error {
	select {
	case c.messageCh <- data:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("send message 5s timeout")
	}
}
