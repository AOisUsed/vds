package aggregator

import (
	"log"
	"sync"
	"virturalDevice/internal/vds/virtualdevice/message"
)

// Aggregator vds中消息集合器，聚合消息并发送到分发器
type Aggregator struct {
	outgoingCh chan message.Task
	wg         sync.WaitGroup
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		outgoingCh: make(chan message.Task, 50),
	}
}

// Watch 添加接收本vds虚拟设备消息的通道,并立刻开始接收消息并处理，上游消息通道关闭，自动停止
func (a *Aggregator) Watch(incomingCh <-chan message.Task) {
	a.wg.Add(1)
	go a.aggregateSingle(incomingCh)
}

// OutChan 消息出口
func (a *Aggregator) OutChan() <-chan message.Task {
	return a.outgoingCh
}

// Stop 停止消息集合器服务 (不会强制停止，而会等待所有上游通道关闭后才停止)
func (a *Aggregator) Stop() {
	a.wg.Wait()
	close(a.outgoingCh)
	log.Println("aggregator 停止")
}

// aggregateSingle 接收特定消息渠道的消息，并发送到统一出口，如果上游通道关闭，则停止监听
func (a *Aggregator) aggregateSingle(incomingCh <-chan message.Task) {
	defer a.wg.Done()
	// log.Println("aggregator 开始监听此通道")
	for msgTask := range incomingCh {
		//log.Printf("出站路由正在将消息转送到统一出口\n")
		select {
		case <-msgTask.Ctx.Done():
			continue
		case a.outgoingCh <- msgTask:
		}
	}
	// log.Println("aggregator 停止监听此通道")
}
