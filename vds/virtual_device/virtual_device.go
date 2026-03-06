package virtual_device

import (
	"log"
	"virturalDevice/cipher"
	"virturalDevice/message"
)

type VirtualDevice struct {
	ID         string
	cipher     cipher.Cipher          // 密码机
	receiveCh  <-chan message.Message // 消息接收通道
	sendCh     chan message.Task      // 消息任务发送通道
	attributes Flag                   // 电台参数
}

func NewVirtualDevice(id string, cipher cipher.Cipher, receiveCh <-chan message.Message) *VirtualDevice {
	return &VirtualDevice{
		ID:        id,
		cipher:    cipher,
		receiveCh: receiveCh,
		sendCh:    make(chan message.Task),
	}
}

// Start 运行虚拟设备，接收消息，打印到控制台
func (vd *VirtualDevice) Start() {
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
	msg := message.Task{
		SrcID:   vd.ID,
		DstID:   dstId,
		Payload: bodyEncrypted,
	}
	vd.sendCh <- msg
}

// OutChan 消息出口
func (vd *VirtualDevice) OutChan() <-chan message.Task {
	return vd.sendCh
}

// Stop 停止虚拟设备，不再发送和接收消息
func (vd *VirtualDevice) Stop() {
	close(vd.sendCh)
}
