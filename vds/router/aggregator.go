package router

import (
	"virturalDevice/message"
)

// Aggregator vds中消息集合器，聚合消息并发送到分发器
type Aggregator struct {
	incomingChs []<-chan message.Task
	outgoingCh  chan<- message.Task
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

// Serve 启动消息出站路由服务
func (a *Aggregator) Serve() {
	defer close(a.outgoingCh)
	for _, ch := range a.incomingChs {
		go a.Aggregate(ch)
	}
}

// Aggregate 接收特定消息渠道的消息，并发送到统一出口，如果消息上游取消了，则取消发送
func (a *Aggregator) Aggregate(incomingCh <-chan message.Task) {
	for msgTask := range incomingCh {
		//log.Printf("出站路由正在将消息转送到统一出口\n")
		select {
		case <-msgTask.Ctx.Done():
			continue
		case a.outgoingCh <- msgTask:
		}
	}
}
