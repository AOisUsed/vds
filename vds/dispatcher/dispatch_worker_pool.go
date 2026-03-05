package dispatcher

import (
	"sync"
	"virturalDevice/message"
)

// DispatchWorkerPool 消息分发工作池
type DispatchWorkerPool struct { // todo:未完成！
	incomingCh <-chan message.Message
	numWorkers int
	wg         *sync.WaitGroup
	done       chan struct{}
}

func NewDispatchWorkerPool(incomingCh <-chan message.Message, numWorkers int) *DispatchWorkerPool {
	return &DispatchWorkerPool{
		incomingCh: incomingCh,
		numWorkers: numWorkers,
		wg:         &sync.WaitGroup{},
		done:       make(chan struct{}),
	}
}
