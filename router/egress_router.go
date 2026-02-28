package router

import (
	"virturalDevice/message"
)

// EgressRouter vds中接受所有所有虚拟设备消息，并聚合发送出去的聚合器
type EgressRouter struct {
	incomingCh []<-chan message.Message
	outgoingCh chan<- message.Message
}

func NewEgressRouter(inputChs []<-chan message.Message, outputCh chan<- message.Message) *EgressRouter {
	return &EgressRouter{
		incomingCh: inputChs,
		outgoingCh: outputCh,
	}
}

// 添加接收本vds虚拟设备消息的通道
func (a *EgressRouter) AddIncomingCh(inputCh <-chan message.Message) {
	a.incomingCh = append(a.incomingCh, inputCh)
}

// Serve 启动消息出站路由服务
func (a *EgressRouter) Serve() {
	for _, ch := range a.incomingCh {
		go a.Route(ch)
	}
}

// Route 接收特定消息渠道的消息，并发送到统一出口
func (a *EgressRouter) Route(ch <-chan message.Message) {
	for msg := range ch {
		//log.Printf("出站路由正在将消息转送到统一出口\n")
		a.outgoingCh <- msg
	}
}
