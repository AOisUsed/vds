package dispatcher

import (
	"log"
	"strconv"
	"testing"
	"time"
	message2 "virturalDevice/internal/vds/virtualdevice/message"
)

// 测试使用Stop()来停止
func TestWorkerPoolStop(t *testing.T) {
	// 任务通道
	inCh := make(chan message2.Task)

	// 任务函数
	handler := func(task message2.Task) {
		time.Sleep(time.Millisecond * 100)
		log.Printf("task %v done\n", task.Message.SrcID)
	}

	// 发送100个任务，并关闭任务通道
	go func() {
		for i := 0; i < 100; i++ {
			inCh <- message2.Task{
				Message: message2.Message{
					SrcID: strconv.Itoa(i),
				},
			}
		}
		log.Printf("all tasks sent\n")
		close(inCh) // 关闭通道以表示没有更多的任务指派，worker自然结束退出
	}()

	// 创建新的工作池，开始接收任务并执行
	wp := NewDispatchWorkerPool(inCh, 3)
	wp.Start(handler)
	time.Sleep(time.Second * 2)
	log.Printf("after 2 seconds, stop worker pool\n")
	wp.Stop()

	// 创建新的工作池，继续未完成任务
	wp = NewDispatchWorkerPool(inCh, 4)
	wp.Start(handler)
	wp.Wait()

}

func TestWorkerPoolWait(t *testing.T) {
	// 任务通道
	inCh := make(chan message2.Task)

	// 任务函数
	handler := func(task message2.Task) {
		time.Sleep(time.Millisecond * 100)
		log.Printf("task %v done\n", task.Message.SrcID)
	}

	wp := NewDispatchWorkerPool(inCh, 3)

	// 发送100个任务，并关闭任务通道
	go func() {
		for i := 0; i < 100; i++ {
			inCh <- message2.Task{
				Message: message2.Message{
					SrcID: strconv.Itoa(i),
				},
			}
		}
		log.Printf("all tasks sent\n")
		close(inCh) // 关闭通道以表示没有更多的任务指派，worker自然结束退出
	}()
	wp.Start(handler)
	wp.Wait() // 等待所有任务完成再结束此进程
}
