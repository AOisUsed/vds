package connection

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
)

type MockConn struct {
	dataCh    chan []byte
	closed    atomic.Bool // 原子标志，替代 sync.Once 的状态检查
	closeOnce sync.Once   // 确保只关闭一次
}

func NewMockConn() *MockConn {
	return &MockConn{
		dataCh: make(chan []byte, 10),
	}
}

func (c *MockConn) Send(ctx context.Context, data []byte) error {
	if c.closed.Load() {
		return nil
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	case c.dataCh <- data:
		return nil
		//case <-time.After(time.Millisecond * 300):
		//	return errors.New("连接阻塞无法发送消息")
	}

}

func (c *MockConn) Receive(ctx context.Context) ([]byte, error) {
	if c.closed.Load() {
		// 检查是否还有数据可读
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		case data, ok := <-c.dataCh:
			if !ok {
				return nil, io.EOF
			}
			return data, nil
		default:
			return nil, io.EOF
		}
	}

	select {
	case data, ok := <-c.dataCh:
		if !ok {
			return nil, io.EOF
		}
		return data, nil
	}
}

func (c *MockConn) Close() error {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		close(c.dataCh)
	})
	return nil
}
