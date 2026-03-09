package virtualdevice

import (
	"context"
	"log"
	"virturalDevice/pkg/cipher"
	"virturalDevice/pkg/message"
	"virturalDevice/pkg/vds/attribute"
)

// VirtualDevice 虚拟通信设备，默认操作是单线程，无并发的
//
// todo: 需要仔细考虑设备的生命周期如何控制
type VirtualDevice struct {
	ID         string
	cipher     cipher.Cipher          // 密码机
	receiveCh  <-chan message.Message // 消息接收通道
	sendCh     chan message.Task      // 消息任务发送通道
	attributes attribute.Flag         // 电台参数

	CancelMessaging context.CancelFunc // 取消消息发送函数
}

func NewVirtualDevice(id string, cipher cipher.Cipher, receiveCh <-chan message.Message) *VirtualDevice {
	return &VirtualDevice{
		ID:        id,
		cipher:    cipher,
		receiveCh: receiveCh,
		sendCh:    make(chan message.Task, 50),
	}
}

// OutChan 消息任务发送出口
func (vd *VirtualDevice) OutChan() <-chan message.Task {
	return vd.sendCh
}

// Run 运行虚拟设备，接收消息，打印到控制台，生命周期由上游关闭通道结束
func (vd *VirtualDevice) Run() {
	//log.Printf("正在运行虚拟设备%v\n", vd.ID)
	for incomingMessage := range vd.receiveCh {
		// 解密消息
		bodyDecrypted, err := vd.cipher.Decrypt(incomingMessage.Payload)
		if err != nil {
			log.Printf("%v无法解密收到的消息: %s\n", vd.ID, err)
			continue
		}

		// 打印消息内容
		log.Printf("虚拟设备 %v 收到消息，内容是：%s\n", vd.ID, bodyDecrypted)
	}
}

// Send 虚拟设备发出消息 (非并发安全)
func (vd *VirtualDevice) Send(dstId string, body []byte) {
	log.Printf("虚拟设备 %v 正在发送消息给%v\n", vd.ID, dstId)
	bodyEncrypted, err := vd.cipher.Encrypt(body)
	if err != nil {
		log.Printf("%v无法加密消息：: %s\n", vd.ID, err)
	}
	msg := message.Message{
		SrcID:   vd.ID,
		DstID:   dstId,
		Payload: bodyEncrypted,
	}
	ctx, cancel := context.WithCancel(context.Background())
	vd.CancelMessaging = cancel
	sendTask := message.Task{
		Ctx:     ctx,
		Message: msg,
	}

	vd.sendCh <- sendTask
}

// CancelSend 取消发送当前正在发送的消息 (非并发安全)
func (vd *VirtualDevice) CancelSend() {
	vd.CancelMessaging()
}

// StopSend 停止虚拟设备发送消息 (非并发安全)
func (vd *VirtualDevice) StopSend() {
	close(vd.sendCh)
}
