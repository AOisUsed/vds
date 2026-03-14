package dispatcher

import (
	"sync"
	"virturalDevice/internal/vds/virtualdevice/message"
)

// WorkerPool 消息分发工作池
type WorkerPool struct {
	incomingTaskCh <-chan message.Task
	numWorkers     int
	wg             sync.WaitGroup
	done           chan struct{} // 通过关闭此通道以停止所有工作
}

func NewDispatchWorkerPool(incomingCh <-chan message.Task, numWorkers int) *WorkerPool {
	return &WorkerPool{
		incomingTaskCh: incomingCh,
		numWorkers:     numWorkers,
		done:           make(chan struct{}),
	}
}

// worker 实际执行dispatch的工作协程
func (wp *WorkerPool) worker(handler func(task message.Task)) {
	defer wp.wg.Done()

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

// Start 启动所有worker(非阻塞)
func (wp *WorkerPool) Start(handler func(task message.Task)) {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(handler)
	}
}

// Wait 等待所有worker完成任务再退出，在主线程中调用以防止worker协程未结束任务就因主线程退出被杀死
//
// 使用方法：
//
//		func main (){
//	    // ... 一些初始化
//		   wp.Start()
//		   // ... 一些操作
//		   wp.Wait() // <- 不能用goroutine执行，要在主线程执行
//		}
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

// Stop 通知所有worker完成当前任务后停止工作（无论上游是否还有持续的任务发送）
//
// 注意：调用Stop()会使上游任务发送通道阻塞, 且Stop()后不可用Start()复用此工作池.
// 如果需要继续任务，创建新的工作池
func (wp *WorkerPool) Stop() {
	close(wp.done)
	wp.wg.Wait()
}
