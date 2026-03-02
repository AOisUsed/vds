package router

import (
	"virturalDevice/message"
	"virturalDevice/vds/registry_syncer"
)

type Dispatcher struct {
	incomingCh     <-chan message.Message          // 出站消息统一出口
	registrySyncer *registry_syncer.RegistrySyncer // 本地注册同步器
}

func NewDispatcher(incomingCh <-chan message.Message, registrySyncer *registry_syncer.RegistrySyncer) *Dispatcher {
	return &Dispatcher{
		incomingCh:     incomingCh,
		registrySyncer: registrySyncer,
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
	sendCh := d.registrySyncer.GetSendChan(incomingMsg.DstID)
	sendCh <- incomingMsg
}
