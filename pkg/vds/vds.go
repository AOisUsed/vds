package vds

import (
	"context"
	"errors"
	"log"
	"runtime"
	"virturalDevice/pkg/cipher"
	"virturalDevice/pkg/connection"
	"virturalDevice/pkg/message"
	"virturalDevice/pkg/vds/aggregator"
	"virturalDevice/pkg/vds/dispatcher"
	"virturalDevice/pkg/vds/ingressrouter"
	"virturalDevice/pkg/vds/sender"
	"virturalDevice/pkg/vds/vdrepository"
	"virturalDevice/pkg/vds/virtualdevice"
)

type VDS struct {
	// 设备
	vdById map[string]*virtualdevice.VirtualDevice // 设备id-实体映射

	// vds 消息入口
	conn       connection.Connection
	incomingCh chan message.Message

	// 数据处理
	vdRepository vdrepository.VDRepository // 虚拟设备信息仓库

	// 消息处理
	ingressRouter *ingressrouter.IngressRouter // 消息入口路由
	aggregator    *aggregator.Aggregator       // 消息集合器
	dispatcher    *dispatcher.Dispatcher       //消息分发器
	sender        sender.Sender                //消息出口发送器
}

// NewVDS 初始化VDS(无设备)
func NewVDS(conn connection.Connection, repo vdrepository.VDRepository, sender sender.Sender) *VDS {
	vds := &VDS{
		vdById:       make(map[string]*virtualdevice.VirtualDevice),
		conn:         conn,
		incomingCh:   make(chan message.Message, 100),
		vdRepository: repo,
		sender:       sender,
	}
	vds.ingressRouter = ingressrouter.NewIngressRouter(vds.incomingCh)
	vds.aggregator = aggregator.NewAggregator()
	vds.dispatcher = dispatcher.NewDispatcher(vds.aggregator.OutChan(), vds.vdRepository, sender, runtime.NumCPU()*2)
	return vds
}

// RegisterDevice 注册vd到vds中 (可在vds运行中随时调用) //todo: 需要考虑失败的处理
func (vds *VDS) RegisterDevice(ctx context.Context, id string, cipher cipher.Cipher) {

	routerOutCh := vds.ingressRouter.CreateOutboundChByID(id)
	vd := virtualdevice.NewVirtualDevice(id, cipher, routerOutCh)
	vdOutCh := vd.OutChan()
	vds.aggregator.Watch(vdOutCh)

	// 调用 vdRepository 把vd注册到registry中
	err := vds.vdRepository.SetVDConnById(ctx, id, vds.conn)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("注册设备失败：%v\n", err)
		// todo: 删除设备

		return
	}
}

// DeleteDevice 移除本地设备(包括ingressrouter, aggregator 相关的引用)
func (vds *VDS) DeleteDevice(id string) {

}

func (vds *VDS) DeregisterDevice(id string) {
	vds.ingressRouter.RemoveOutboundChByID(id)
	delete(vds.vdById, id)

}

// Run 启动vds服务
func (vds *VDS) Run() {
	//log.Println(" VDS 正在启动")

}
