package router

import (
	"log"
	"virturalDevice/message"
)

// IngressRouter vds中的入站路由，可以将vds收到的消息发送到对应的虚拟设备中
type IngressRouter struct {
	inboundCh   <-chan message.Message
	outboundChs map[string]chan<- message.Message
}

func NewIngressRouter(inboundCh <-chan message.Message, outboundChs map[string]chan<- message.Message) *IngressRouter {
	return &IngressRouter{
		inboundCh:   inboundCh,
		outboundChs: outboundChs,
	}
}

// AddOutboundCh 添加消息路由去向通道
func (r *IngressRouter) AddOutboundCh(id string, outboundCh chan<- message.Message) {
	r.outboundChs[id] = outboundCh
}

func (r *IngressRouter) InboundCh() <-chan message.Message {
	return r.inboundCh
}

// Serve 启动入站路由服务
func (r *IngressRouter) Serve() {
	//log.Println("正在启动入站路由")
	for msg := range r.inboundCh {
		r.Route(msg)
	}
}

// Route 根据消息中dstID进行路由（简易实现）
func (r *IngressRouter) Route(msg message.Message) {
	dstId := msg.DstID
	if r.outboundChs[dstId] != nil {
		// log.Printf("路由器正在将消息发送至目标ID为%v的设备\n", dstId)
		r.outboundChs[dstId] <- msg
	} else {
		log.Printf("路由表内找不到目标ID为%v的设备\n", dstId)
	}
}
