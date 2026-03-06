package vds

import (
	"sync"
	"virturalDevice/message"
)

// Aggregator vds中消息集合器，聚合消息并发送到分发器
type Aggregator struct {
	incomingChs []<-chan message.Task
	outgoingCh  chan<- message.Task
	wg          *sync.WaitGroup
}

func NewAggregator(incomingCh []<-chan message.Task, outgoingCh chan<- message.Task) *Aggregator {
	return &Aggregator{
		incomingChs: incomingCh,
		outgoingCh:  outgoingCh,
	}
}

// AddIncomingCh 添加接收本vds虚拟设备消息的通道
func (a *Aggregator) AddIncomingCh(incomingCh <-chan message.Task) {
	a.incomingChs = append(a.incomingChs, incomingCh)
}

// Serve 启动消息集合器服务
func (a *Aggregator) Serve() {
	for _, ch := range a.incomingChs {
		a.wg.Add(1)
		go a.Aggregate(ch)
	}
}

// Stop 停止消息集合器服务
func (a *Aggregator) Stop() {
	a.wg.Wait()
	close(a.outgoingCh)
}

// Aggregate 接收特定消息渠道的消息，并发送到统一出口，如果消息上游取消了，则取消发送
func (a *Aggregator) Aggregate(incomingCh <-chan message.Task) {
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
