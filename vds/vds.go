package vds

import (
	"context"
	"log"
	"virturalDevice/message"
	"virturalDevice/vds/connection"
	"virturalDevice/vds/dispatcher"
	"virturalDevice/vds/repository"
	"virturalDevice/vds/router"
	"virturalDevice/vds/sender"
	"virturalDevice/vds/virtual_device"
)

type VDS struct {
	// 设备
	deviceById map[string]*virtual_device.VirtualDevice // 设备id-实体映射

	// vds 地址
	address connection.Conn

	// 统一消息出入口
	inputCh  chan message.Message // 整体消息入口
	outputCh chan message.Message // 整体消息出口

	// 消息处理
	ingressRouter *router.IngressRouter  // 消息入口路由
	aggregator    *router.Aggregator     // 消息集合器
	dispatcher    *dispatcher.Dispatcher //消息分发器
	sender        sender.Sender          //消息出口发送器

	// 数据处理
	vdRepository repository.VDRepository // 虚拟设备信息仓库
}

// NewVDS 初始化VDS(无设备)
func NewVDS(inputCh chan message.Message, outputCh chan message.Message, vdRepository repository.VDRepository, sender sender.Sender, address connection.Conn) *VDS {
	vds := &VDS{
		deviceById:    make(map[string]*virtual_device.VirtualDevice),
		address:       address,
		inputCh:       inputCh,
		outputCh:      outputCh,
		ingressRouter: router.NewIngressRouter(inputCh, make(map[string]chan<- message.Message)),
		aggregator:    router.NewAggregator([]<-chan message.Message{}, outputCh),
		dispatcher:    dispatcher.NewDispatcher(outputCh, vdRepository, sender),
		sender:        sender,
		vdRepository:  vdRepository,
	}

	return vds
}

// RegisterDevice 注册设备到vds中，同时将设备的输入输出通道添加到vds的进站路由器和出站路由器中，再通过调用registrySyncer把设备注册到注册中心中
func (vds *VDS) RegisterDevice(device *virtual_device.VirtualDevice) {
	// 本地注册 vd 到设备列表，并加入进站出站路由中
	vds.deviceById[device.ID] = device
	vds.ingressRouter.AddOutboundCh(device.ID, device.ReceiveChan())
	vds.aggregator.AddIncomingCh(device.SendChan())

	// 调用 vdRepository 把vd注册到registry中
	err := vds.vdRepository.SetVDConnById(context.Background(), device.ID, vds.address) // 暂时使用默认context， 后续有控制注册vd生命周期再修改
	if err != nil {
		log.Println(err)
		return
	}
}

// Serve 启动vds服务
func (vds *VDS) Serve() {
	//log.Println(" VDS 正在启动")
	for _, device := range vds.deviceById {
		go device.Run()
	}
	go vds.ingressRouter.Serve()
	go vds.aggregator.Serve()
	go vds.dispatcher.Serve()
}
