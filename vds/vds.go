package vds

import (
	"context"
	"log"
	"runtime"
	"virturalDevice/cipher"
	"virturalDevice/connection"
	"virturalDevice/message"
	"virturalDevice/vds/virtual_device"
)

type VDS struct {
	// 设备
	vdById map[string]*virtual_device.VirtualDevice // 设备id-实体映射

	// vds 消息入口
	conn       connection.Connection
	incomingCh chan message.Message

	// 数据处理
	vdRepository VDRepository // 虚拟设备信息仓库

	// 消息处理
	ingressRouter *IngressRouter // 消息入口路由
	aggregator    *Aggregator    // 消息集合器
	dispatcher    *Dispatcher    //消息分发器
	sender        Sender         //消息出口发送器
}

// NewVDS 初始化VDS(无设备)
func NewVDS(conn connection.Connection, repo VDRepository, sender Sender) *VDS {
	vds := &VDS{
		vdById:       make(map[string]*virtual_device.VirtualDevice),
		conn:         conn,
		incomingCh:   make(chan message.Message, 100),
		vdRepository: repo,
		sender:       sender,
	}
	vds.ingressRouter = NewIngressRouter(vds.incomingCh)
	vds.aggregator = NewAggregator()
	vds.dispatcher = NewDispatcher(vds.aggregator.OutChan(), vds.vdRepository, sender, runtime.NumCPU()*2)
	return vds
}

// RegisterDevice 注册vd到vds中 (可在vds运行中随时调用) //todo: 需要考虑失败的处理
func (vds *VDS) RegisterDevice(id string, cipher cipher.Cipher) {

	routerOutCh := vds.ingressRouter.CreateOutboundChByID(id)
	vd := virtual_device.NewVirtualDevice(id, cipher, routerOutCh)
	vdOutCh := vd.OutChan()
	vds.aggregator.AddIncomingCh(vdOutCh)

	// 调用 vdRepository 把vd注册到registry中
	err := vds.vdRepository.SetVDConnById(context.Background(), id, vds.conn) // 暂时使用默认context， 后续有控制注册vd生命周期再修改
	if err != nil {
		log.Println(err)
		return
	}
}

func (vds *VDS) DeregisterDevice(device *virtual_device.VirtualDevice) {

}

// Run 启动vds服务
func (vds *VDS) Run() {
	//log.Println(" VDS 正在启动")
	for _, device := range vds.vdById {
		go device.Start()
	}
	go vds.ingressRouter.Run()
	go vds.aggregator.Run()
	go vds.dispatcher.Run()
}
