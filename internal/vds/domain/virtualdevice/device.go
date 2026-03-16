package virtualdevice

import (
	"context"
	"log"
	"sync"
	"virturalDevice/internal/vds/domain/message"
	"virturalDevice/internal/vds/domain/virtualdevice/cipher"
	"virturalDevice/internal/vds/domain/virtualdevice/params"
)

// VirtualDevice 虚拟通信设备，默认操作是单线程，并发不安全
type VirtualDevice struct {
	id        string
	cipher    cipher.Cipher          // 密码机
	receiveCh <-chan message.Message // 消息接收通道
	sendCh    chan message.Task      // 消息任务发送通道
	params    params.Params          // 设备参数

	subscriptionCh chan message.Message // 暴露给外部的观察设备接收消息的通道

	cancelMessaging context.CancelFunc // 取消消息发送函数
	stop            chan struct{}      // 停止工作信号通道

	rwMu sync.RWMutex
}

type Option func(*VirtualDevice)

func WithCipher(c cipher.Cipher) Option {
	return func(vd *VirtualDevice) {
		vd.cipher = c
	}
}

func WithParams(p params.Params) Option {
	return func(vd *VirtualDevice) {
		vd.params = p
	}
}

func NewVirtualDevice(id string, receiveCh <-chan message.Message, opts ...Option) *VirtualDevice {
	vd := &VirtualDevice{
		id:        id,
		receiveCh: receiveCh,
		sendCh:    make(chan message.Task, 50), // 缓存大小可以根据实际情况调整，暂时设为50
		stop:      make(chan struct{}),

		subscriptionCh: make(chan message.Message),

		cipher: cipher.NewPlain(), // 默认明文密码机
		params: params.NewEmpty(), // 默认空白参数
	}

	for _, opt := range opts {
		opt(vd)
	}

	return vd
}

// Params 设备参数
func (vd *VirtualDevice) Params() params.Params {
	vd.rwMu.RLock()
	defer vd.rwMu.RUnlock()
	return vd.params
}

// OutChan 获取消息任务发送出口
func (vd *VirtualDevice) OutChan() <-chan message.Task {
	return vd.sendCh
}

// SendMessage 虚拟设备发出消息
//
// dstId 为空进行广播，不为空进行单播
func (vd *VirtualDevice) SendMessage(dstId string, body []byte) {
	vd.rwMu.Lock()
	defer vd.rwMu.Unlock()

	if dstId == "" {
		log.Printf("虚拟设备 %v 正在广播消息\n", vd.id)
	} else {
		log.Printf("虚拟设备 %v 正在发送消息给%v\n", vd.id, dstId)
	}

	bodyEncrypted, err := vd.cipher.Encrypt(body)
	if err != nil {
		log.Printf("%v无法加密消息: %s\n", vd.id, err)
	}
	msg := message.Message{
		SrcID:   vd.id,
		DstID:   dstId,
		Payload: bodyEncrypted,
	}
	ctx, cancel := context.WithCancel(context.Background())
	vd.cancelMessaging = cancel
	sendTask := message.Task{
		Ctx:     ctx,
		Message: msg,
	}

	vd.sendCh <- sendTask
}

// CancelSend 取消发送当前正在发送的消息 (非并发安全)
func (vd *VirtualDevice) CancelSend() {
	vd.rwMu.Lock()
	defer vd.rwMu.Unlock()

	if vd.cancelMessaging != nil {
		vd.cancelMessaging()
	} else {
		log.Printf("虚拟设备 %v 当前无正在发送的消息，无法取消\n", vd.id)
	}
}

// SubscribeMessage 订阅
func (vd *VirtualDevice) SubscribeMessage() <-chan message.Message {
	vd.rwMu.RLock()
	defer vd.rwMu.RUnlock()

	return vd.subscriptionCh
}

// Run 运行虚拟设备，接收消息，打印到控制台，生命周期由上游关闭通道结束
func (vd *VirtualDevice) Run() {
	log.Printf("开始运行设备%v\n", vd.id)
	defer log.Printf("设备%v停止运行\n", vd.id)

	for {
		select {
		case <-vd.stop:
			return
		case incomingMessage, ok := <-vd.receiveCh:
			if !ok {
				return
			}
			// 解密消息
			bodyDecrypted, err := vd.cipher.Decrypt(incomingMessage.Payload)
			if err != nil {
				log.Printf("%v无法解密收到的消息: %s\n", vd.id, err)
				continue
			}
			// 打印消息内容
			log.Printf("虚拟设备 %v 收到消息，内容是：%q\n", vd.id, bodyDecrypted)

			// 把消息转送给接收消息的订阅者(如接收消息的前端)，无人观察则直接丢弃消息
			select {
			case vd.subscriptionCh <- incomingMessage:
			default:
			}
		}
	}
}

// Stop 停止虚拟设备发送/接收消息 (会阻塞上游receiveCh) (非并发安全) // todo: 如果后续添加的其他业务会启动常驻goroutine，也要添加对应的资源关闭机制
func (vd *VirtualDevice) Stop() {
	close(vd.stop)
	close(vd.sendCh)
}
