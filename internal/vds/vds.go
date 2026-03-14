package vds

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"runtime"
	"sync"
	"time"
	"virturalDevice/internal/vds/aggregator"
	"virturalDevice/internal/vds/codec"
	"virturalDevice/internal/vds/connection"
	"virturalDevice/internal/vds/dispatcher"
	"virturalDevice/internal/vds/ingressrouter"
	"virturalDevice/internal/vds/repository"
	"virturalDevice/internal/vds/sender"
	"virturalDevice/internal/vds/virtualdevice"
	"virturalDevice/internal/vds/virtualdevice/message"
)

type VDS struct {
	// 设备
	vdById map[string]*virtualdevice.VirtualDevice // 设备id-实体映射

	// vds 消息入口
	conn       connection.Connection
	incomingCh chan message.Message

	// 数据
	vdRepository repository.VDRepository // 虚拟设备信息仓库

	// 消息处理
	ingressRouter *ingressrouter.IngressRouter // 消息入口路由
	aggregator    *aggregator.Aggregator       // 消息集合器
	dispatcher    *dispatcher.Dispatcher       //消息分发器
	sender        sender.Sender                //消息出口发送器

	// 编解码器
	codec codec.Codec

	rwMutex sync.RWMutex
}

// NewVDS 初始化VDS(无设备)
func NewVDS(conn connection.Connection, repo repository.VDRepository, sender sender.Sender, codec codec.Codec) *VDS {
	vds := &VDS{
		vdById:       make(map[string]*virtualdevice.VirtualDevice),
		conn:         conn,
		incomingCh:   make(chan message.Message, 100), // 设置100的缓存
		vdRepository: repo,
		sender:       sender,
		codec:        codec,
	}
	vds.ingressRouter = ingressrouter.NewIngressRouter(vds.incomingCh)
	vds.aggregator = aggregator.NewAggregator()
	vds.dispatcher = dispatcher.NewDispatcher(vds.aggregator.OutChan(), vds.vdRepository, vds.codec, sender, runtime.NumCPU()*2)
	return vds
}

func (vds *VDS) Device(id string) *virtualdevice.VirtualDevice {
	vds.rwMutex.RLock()
	defer vds.rwMutex.RUnlock()

	v, ok := vds.vdById[id]
	if !ok {
		return nil
	}
	return v
}

// readerLoop 持续从连接读取数据
func (vds *VDS) readerLoop() {
	defer close(vds.incomingCh)

	for {
		data, err := vds.conn.Receive()
		if err != nil {
			if err != io.EOF {
				log.Printf("无法从连接读取数据:%v \n", err)

				time.Sleep(3 * time.Second) // 出现EOF外其他error后, 5秒后再重新尝试读数据
				continue
			}
			log.Printf("连接关闭，停止从中读取数据\n")
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

// UpdateDeviceParams 数据仓库中更新设备参数
func (vds *VDS) UpdateDeviceParams(ctx context.Context, id string) error {
	vds.rwMutex.RLock()

	vd, ok := vds.vdById[id]
	if !ok {
		vds.rwMutex.RUnlock()
		return errors.New(fmt.Sprintf("vds中无设备%v", id))
	}
	vds.rwMutex.RUnlock()

	params := vd.Params()
	err := vds.vdRepository.SetParams(ctx, id, params)
	return err
}

// RegisterDeviceConn 数据仓库中添加设备连接信息
func (vds *VDS) RegisterDeviceConn(ctx context.Context, id string) error {
	err := vds.vdRepository.SetConnection(ctx, id, vds.conn)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Printf("注册设备%v连接信息取消：%v\n", id, err)
		} else {
			log.Printf("注册设备%v连接信息失败：%v\n", id, err)
		}
	}
	return err
}

// DeregisterDeviceConn 删除数据仓库中设备连接信息
func (vds *VDS) DeregisterDeviceConn(ctx context.Context, id string) error {
	err := vds.vdRepository.RemoveConnection(ctx, id)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Printf("删除设备%v连接信息取消：%v\n", id, err)
		} else {
			log.Printf("删除设备%v连接信息失败：%v\n", id, err)
		}
	}
	return err
}

// connectDevice vds中创建设备，与前后消息通道连接，并运行设备，如果同id的设备存在会删除原设备
func (vds *VDS) connectDevice(id string, opts ...virtualdevice.Option) {
	vds.rwMutex.Lock()
	defer vds.rwMutex.Unlock()

	if _, exists := vds.vdById[id]; exists {
		vds.disconnectDevice(id)
	}
	routerOutCh := vds.ingressRouter.CreateOutboundCh(id)
	vd := virtualdevice.NewVirtualDevice(id, routerOutCh, opts...)
	vdOutCh := vd.OutChan()
	vds.aggregator.Watch(vdOutCh)
	go vd.Run()
	vds.vdById[id] = vd
}

// disconnectDevice 停止并移除设备和及相关消息收发通道
func (vds *VDS) disconnectDevice(id string) {
	vds.rwMutex.Lock()
	defer vds.rwMutex.Unlock()

	if _, exists := vds.vdById[id]; !exists {
		return
	}
	vds.ingressRouter.RemoveOutboundCh(id)
	vds.vdById[id].Stop()
	delete(vds.vdById, id)
}

// ConnectAndRegisterDevice 连接并注册设备连接信息到数据库中
func (vds *VDS) ConnectAndRegisterDevice(ctx context.Context, id string, opts ...virtualdevice.Option) error {

	// 创建并运行设备
	vds.connectDevice(id, opts...)

	// 注册设备到数据仓库中
	err := vds.RegisterDeviceConn(ctx, id)
	if err != nil {
		// 失败则回滚
		vds.disconnectDevice(id)
		return err
	}
	log.Printf("注册设备%v连接信息成功\n", id)
	return nil
}

// DisconnectAndDeregisterDevice 停止，断开vd设备，并删除数据库中设备连接信息
func (vds *VDS) DisconnectAndDeregisterDevice(ctx context.Context, id string) error {

	err := vds.DeregisterDeviceConn(ctx, id)
	if err != nil {
		return err
	}
	log.Printf("删除设备%v连接信息成功\n", id)
	vds.disconnectDevice(id)
	return err
}

//
//func (vds *VDS) UpdateDeviceParams(ctx context.Context, id string) error{
//	vds.rwMutex.Lock()
//	defer vds.rwMutex.Unlock()
//}

// LoadDevices 从repository载入并运行设备
//func (vds *VDS) LoadDevices(ctx context.Context) error {
//
//}

// Start 启动vds服务
func (vds *VDS) Start() {
	//log.Println("正在启动 vds ")
	go vds.dispatcher.Run()
	go vds.ingressRouter.Run()
	go vds.readerLoop()
}

// Stop 停止vds服务
func (vds *VDS) Stop() {
	vds.rwMutex.Lock()
	defer vds.rwMutex.Unlock()

	err := vds.conn.Close()
	if err != nil {
		log.Printf("连接关闭失败:%v \n", err)
	}

	vds.ingressRouter.Stop()
	for id, vd := range vds.vdById {
		go func(id string, vd *virtualdevice.VirtualDevice) {
			_ = vds.DeregisterDeviceConn(context.Background(), id) // 尝试删除数据仓库中设备连接信息，删除失败也要停止本地goroutine
			vd.Stop()
		}(id, vd)

	}
	vds.aggregator.Stop()
	vds.dispatcher.Stop()

	log.Println("vds 停止")
}
