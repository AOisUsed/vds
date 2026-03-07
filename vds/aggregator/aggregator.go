package aggregator

import (
	"sync"
	"virturalDevice/message"
)

// Aggregator vds中消息集合器，聚合消息并发送到分发器
type Aggregator struct {
	outgoingCh chan message.Task
	wg         *sync.WaitGroup
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		outgoingCh: make(chan message.Task),
	}
}

// AddIncomingCh 添加接收本vds虚拟设备消息的通道,并立刻开始收听消息
func (a *Aggregator) AddIncomingCh(incomingCh <-chan message.Task) {
	a.wg.Add(1)
	go a.aggregateSingle(incomingCh)
}

// OutChan 消息出口
func (a *Aggregator) OutChan() <-chan message.Task {
	return a.outgoingCh
}

// Run 启动消息集合器服务(适用于有初始通道需要监听的情况，初始如果没有监听的通道，不需要调用此方法)
func (a *Aggregator) Run(initialChs []<-chan message.Task) {
	for _, ch := range initialChs {
		a.aggregateSingle(ch)
	}
}

// Stop 停止消息集合器服务
func (a *Aggregator) Stop() {
	a.wg.Wait()
	close(a.outgoingCh)
}

// aggregateSingle 接收特定消息渠道的消息，并发送到统一出口，如果消息上游取消了，则取消发送
func (a *Aggregator) aggregateSingle(incomingCh <-chan message.Task) {
	defer a.wg.Done()
	for msgTask := range incomingCh {
		//log.Printf("出站路由正在将消息转送到统一出口\n")
		select {
		case <-msgTask.Ctx.Done():
			continue
		case a.outgoingCh <- msgTask:
		}
	}
}
