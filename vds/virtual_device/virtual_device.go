package virtual_device

import (
	"log"
	"virturalDevice/cipher"
	"virturalDevice/message"
)

type VirtualDevice struct {
	ID         string
	cipher     cipher.Cipher        // 密码机
	receiveCh  chan message.Message // 消息接收通道
	sendCh     chan message.Message // 消息发送通道
	attributes Flag                 //电台属性
}

func NewVirtualDevice(id string, cipher cipher.Cipher) *VirtualDevice {
	return &VirtualDevice{
		ID:        id,
		cipher:    cipher,
		receiveCh: make(chan message.Message),
		sendCh:    make(chan message.Message),
	}
}

// ReceiveChan 获得该虚拟设备的接收通道
func (vd *VirtualDevice) ReceiveChan() chan<- message.Message {
	return vd.receiveCh
}

// SendChan 获得该虚拟设备的发送通道
func (vd *VirtualDevice) SendChan() <-chan message.Message {
	return vd.sendCh
}

// Run 运行虚拟设备，接收消息，打印到控制台
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

// Send 虚拟设备发出消息
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
	vd.sendCh <- msg
}

// Stop 停止虚拟设备，不再发送和接收消息
func (vd *VirtualDevice) Stop() {
	close(vd.receiveCh)
	close(vd.sendCh)
}
