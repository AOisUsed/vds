package aggregator

import (
	"context"
	"log"
	"strconv"
	"testing"
	"virturalDevice/internal/vds/message"
)

func TestAggregatorBasicLifeCycle(t *testing.T) {

	// aggregator 消息入口
	inChs := make([]chan message.Task, 3)
	for i := 0; i < 3; i++ {
		inChs[i] = make(chan message.Task)
	}

	a := NewAggregator()
	outCh := a.OutChan()

	// 运行 aggregator
	for _, ch := range inChs {
		a.Watch(ch)
	}

	// 消费 aggregator 出口消息并打印日志
	go func(outCh <-chan message.Task) {
		for task := range outCh {
			log.Printf("task received from %v, message: %v\n", task.Message.SrcID, task.Message.Payload)
		}
	}(outCh)

	// 每个生产者并发给 aggregator 发5条消息
	for i := range inChs {
		go func(i int) {
			for j := 0; j < 5; j++ {
				inChs[i] <- message.Task{
					Ctx: context.Background(),
					Message: message.Message{
						SrcID:   strconv.Itoa(i),
						Payload: []byte{byte(j)},
					},
				}
			}
			close(inChs[i])
		}(i)
	}

	a.Stop()
}
