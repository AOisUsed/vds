package vds

import (
	"sync"
	"virturalDevice/message"
)

// DispatchWorkerPool 消息分发工作池
type DispatchWorkerPool struct { // todo:未完成！
	incomingTaskCh <-chan message.Task
	numWorkers     int
	wg             *sync.WaitGroup
	done           chan struct{}
}

func NewDispatchWorkerPool(incomingCh <-chan message.Task, numWorkers int) *DispatchWorkerPool {
	return &DispatchWorkerPool{
		incomingTaskCh: incomingCh,
		numWorkers:     numWorkers,
		wg:             &sync.WaitGroup{},
		done:           make(chan struct{}),
	}
}

// worker 实际执行dispatch的工作协程
func (wp *DispatchWorkerPool) worker(wg *sync.WaitGroup, handler func(task message.Task)) {
	defer wg.Done()

	for {
		select {
		case <-wp.done:
			return
		case task, ok := <-wp.incomingTaskCh:
			if !ok {
				return
			}
			handler(task)
		}
	}
}
