package vds

import (
	"log"
	"sync"
	"virturalDevice/message"
)

// IngressRouter vds中的入口路由，可以将vds收到的消息发送到对应的虚拟设备中
type IngressRouter struct {
	inboundCh      <-chan message.Message
	outboundChByID map[string]chan message.Message
	mu             sync.Mutex
}

func NewIngressRouter(inboundCh <-chan message.Message) *IngressRouter {
	return &IngressRouter{
		inboundCh:      inboundCh,
		outboundChByID: make(map[string]chan message.Message),
	}
}

// CreateOutboundChByID  通过vdID添加消息路由出口通道
func (r *IngressRouter) CreateOutboundChByID(id string) <-chan message.Message {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.outboundChByID[id] = make(chan message.Message, 100) //设置大小为100的缓冲，应对下游无法接收消息的情况
	return r.outboundChByID[id]
}

// RemoveOutboundChByID  通过vdID删除消息路由出口通道
func (r *IngressRouter) RemoveOutboundChByID(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.outboundChByID, id)
}

// OutChByID 通过vdID获得消息路由出口通道
func (r *IngressRouter) OutChByID(id string) <-chan message.Message {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.outboundChByID[id]
}

// Run 启动入站路由
func (r *IngressRouter) Run() {
	//log.Println("正在启动入站路由")
	for msg := range r.inboundCh {
		r.Route(msg)
	}
}

// Route 根据消息中dstID进行路由（简易实现）
func (r *IngressRouter) Route(msg message.Message) {
	dstId := msg.DstID
	if r.outboundChByID[dstId] != nil {
		// log.Printf("路由器正在将消息发送至目标ID为%v的设备\n", dstId)
		select {
		case r.outboundChByID[dstId] <- msg:
		default:
			log.Printf("%v 设备消息接收通道阻塞, 消息被丢弃\n", dstId)
			//todo: 可能需要考虑更好的处理方法。但vd,ingressRouter生命周期由vds管理，正常情况不会出现丢弃消息。
		}
	} else {
		log.Printf("本地找不到目标ID为%v的设备\n", dstId)
	}
}
