package vds

import (
	"virturalDevice/message"
	"virturalDevice/registry"
	"virturalDevice/router"
	"virturalDevice/virtual_device"
)

type VDS struct {
	devices map[string]*virtual_device.VirtualDevice

	inputCh  chan message.Message // 整体消息入口
	outputCh chan message.Message // 整体消息出口

	ingressRouter *router.IngressRouter // 消息入站路由
	egressRouter  *router.EgressRouter  // 消息出站路由
	dispatcher    *router.Dispatcher    //消息分发器

	registry *registry.VDRegistry // 注册中心
}

// NewVDS 初始化VDS(无设备)
func NewVDS(inputCh chan message.Message, outputCh chan message.Message, registry *registry.VDRegistry) *VDS {
	vds := &VDS{
		devices:       make(map[string]*virtual_device.VirtualDevice),
		inputCh:       inputCh,
		outputCh:      outputCh,
		ingressRouter: router.NewIngressRouter(inputCh, make(map[string]chan<- message.Message)),
		egressRouter:  router.NewEgressRouter([]<-chan message.Message{}, outputCh),
		dispatcher:    router.NewDispatcher(outputCh, registry),
		registry:      registry,
	}
	return vds
}

// RegisterDevice 注册设备到vds中，同时将设备的输入输出通道添加到vds的路由器和聚合器中，再把设备注册到注册中心中
func (vds *VDS) RegisterDevice(device *virtual_device.VirtualDevice) {
	vds.devices[device.ID] = device
	vds.ingressRouter.AddOutboundCh(device.ID, device.ReceiveChan())
	vds.egressRouter.AddIncomingCh(device.SendChan())

	vds.registry.RegisterVD(device.ID, vds.inputCh)
}

// Serve 启动vds服务
func (vds *VDS) Serve() {
	//log.Println(" VDS 正在启动")
	for _, device := range vds.devices {
		go device.Run()
	}
	go vds.ingressRouter.Serve()
	go vds.egressRouter.Serve()
	go vds.dispatcher.Serve()
}
