package mock

import (
	"errors"
	"io"
	"time"
)

type Conn struct {
	dataCh chan []byte
}

func NewConn() *Conn {
	return &Conn{
		dataCh: make(chan []byte, 10),
	}
}

func (c *Conn) Send(data []byte) error {
	select {
	case c.dataCh <- data:
		return nil
	case <-time.After(time.Millisecond * 300):
		return errors.New("连接阻塞无法发送消息")
	}

}

func (c *Conn) Receive() ([]byte, error) {
	select {
	case data, ok := <-c.dataCh:
		if !ok {
			return nil, io.EOF
		}
		return data, nil
	}
}

func (c *Conn) Close() error {
	close(c.dataCh)
	return nil
}
