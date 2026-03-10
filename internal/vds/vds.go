package vds

import (
	"context"
	"errors"
	"log"
	"runtime"
	"sync"
	"virturalDevice/internal/cipher"
	"virturalDevice/internal/connection"
	"virturalDevice/internal/message"
	"virturalDevice/internal/vds/aggregator"
	"virturalDevice/internal/vds/codec"
	"virturalDevice/internal/vds/dispatcher"
	"virturalDevice/internal/vds/ingressrouter"
	"virturalDevice/internal/vds/sender"
	"virturalDevice/internal/vds/types"
	"virturalDevice/internal/vds/vdrepository"
	"virturalDevice/internal/vds/virtualdevice"
)

type VDS struct {
	// 设备
	vdById map[string]*virtualdevice.VirtualDevice // 设备id-实体映射

	// vds 消息入口
	conn       connection.Connection
	incomingCh chan message.Message

	// 数据
	vdRepository vdrepository.VDRepository // 虚拟设备信息仓库

	// 消息处理
	ingressRouter *ingressrouter.IngressRouter // 消息入口路由
	aggregator    *aggregator.Aggregator       // 消息集合器
	dispatcher    *dispatcher.Dispatcher       //消息分发器
	sender        sender.Sender                //消息出口发送器

	// 编解码器
	codec codec.Codec

	stop chan struct{}
	mu   sync.Mutex
}

// NewVDS 初始化VDS(无设备)
func NewVDS(conn connection.Connection, repo vdrepository.VDRepository, sender sender.Sender, codec codec.Codec) *VDS {
	vds := &VDS{
		vdById:       make(map[string]*virtualdevice.VirtualDevice),
		conn:         conn,
		incomingCh:   make(chan message.Message, 100), // 设置100的缓存
		vdRepository: repo,
		sender:       sender,
		codec:        codec,
		stop:         make(chan struct{}),
	}
	vds.ingressRouter = ingressrouter.NewIngressRouter(vds.incomingCh)
	vds.aggregator = aggregator.NewAggregator()
	vds.dispatcher = dispatcher.NewDispatcher(vds.aggregator.OutChan(), vds.vdRepository, vds.codec, sender, runtime.NumCPU()*2)
	return vds
}

// readerLoop 持续从连接读取数据 todo:需要解决生命周期的问题
func (vds *VDS) readerLoop() {
	defer close(vds.incomingCh)
	for {
		data, err := vds.conn.Receive()
		if err != nil {
			log.Printf("无法从连接读取数据:%v \n", err)
			return
		}
		msg, err := vds.codec.Decode(data)
		if err != nil {
			log.Printf("无法解码接收到的数据:%v \n", err)
			return
		}
		vds.incomingCh <- msg
	}
}

// RegisterDevice 注册并运行vd (可在vds运行中随时调用) 注：如果同id的设备已经存在，会覆盖原设备
func (vds *VDS) RegisterDevice(ctx context.Context, id string, cipher cipher.Cipher, params types.VDParams) error {
	vds.mu.Lock()
	defer vds.mu.Unlock()

	// 创建并运行 vd
	vds.addDeviceUnsafe(id, cipher, params)
	// 调用 vdRepository 把vd注册到registry中
	err := vds.vdRepository.SetVDConnById(ctx, id, vds.conn)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Printf("注册设备取消：%v\n", err)
		} else {
			log.Printf("注册设备失败：%v\n", err)
		}
		// 失败，回滚删除本地设备
		vds.removeDeviceUnsafe(id)
		return err
	}
	return err
}

// DeregisterDevice 停止并删除vd连接信息 (可在vds运行中随时调用)
func (vds *VDS) DeregisterDevice(ctx context.Context, id string) error {
	vds.mu.Lock()
	defer vds.mu.Unlock()

	err := vds.vdRepository.RemoveVDConnById(ctx, id)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Printf("删除设备注册信息取消：%v\n", err)
		} else {
			log.Printf("删除设备注册信息失败：%v\n", err)
		}
		return err
	}
	vds.removeDeviceUnsafe(id)
	return err
}

// LoadDevices 从repository载入并运行设备
//func (vds *VDS) LoadDevices(ctx context.Context) error {
//
//}

// addDeviceUnsafe vds内存中添加并运行设备，如果同id的设备存在会删除原设备 (非并发安全)
func (vds *VDS) addDeviceUnsafe(id string, cipher cipher.Cipher, params types.VDParams) {
	if _, exists := vds.vdById[id]; exists {
		vds.removeDeviceUnsafe(id)
	}
	routerOutCh := vds.ingressRouter.CreateOutboundChByID(id)
	vd := virtualdevice.NewVirtualDevice(id, cipher, routerOutCh, params)
	vdOutCh := vd.OutChan()
	vds.aggregator.Watch(vdOutCh)
	go vd.Run()
	vds.vdById[id] = vd
}

// removeDeviceUnsafe vds内存中停止并移除设备 (非并发安全)
func (vds *VDS) removeDeviceUnsafe(id string) {
	if _, exists := vds.vdById[id]; !exists {
		return
	}
	vds.ingressRouter.RemoveOutboundChByID(id)
	vds.vdById[id].Stop()
	delete(vds.vdById, id)
}

// Run 启动vds服务
func (vds *VDS) Run() {
	go vds.dispatcher.Run()
	go vds.ingressRouter.Run()
	go vds.readerLoop()
}

// Stop 停止vds服务
func (vds *VDS) Stop() {
	vds.mu.Lock()
	defer vds.mu.Unlock()

	vds.ingressRouter.Stop()
	for _, vd := range vds.vdById {
		vd.Stop()
	}
	vds.aggregator.Stop()
	vds.dispatcher.Stop()
}
