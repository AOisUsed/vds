package router

import (
	"virturalDevice/message"
	"virturalDevice/registry"
)

type Dispatcher struct {
	incomingCh <-chan message.Message
	vdRegistry *registry.VDRegistry
}

func NewDispatcher(incomingCh <-chan message.Message, registry *registry.VDRegistry) *Dispatcher {
	return &Dispatcher{
		incomingCh: incomingCh,
		vdRegistry: registry,
	}
}

// Serve 运行消息分发器
func (d *Dispatcher) Serve() {
	for msg := range d.incomingCh {
		d.Dispatch(msg)
	}
}

// Dispatch 根据消息中dstId和注册中心中vds消息接收通道分发单条消息到对应vds
func (d *Dispatcher) Dispatch(incomingMsg message.Message) {
	sendCh := d.vdRegistry.GetSendChan(incomingMsg.DstID)
	sendCh <- incomingMsg
}
