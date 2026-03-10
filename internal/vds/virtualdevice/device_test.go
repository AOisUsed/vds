package virtualdevice

import (
	"log"
	"strconv"
	"sync"
	"testing"
	"time"
	"virturalDevice/internal/message"
	"virturalDevice/internal/mock"
)

func TestDeviceSend(t *testing.T) {

	inCh := make(chan message.Message)
	ciph := mock.NewCipher()
	dv := NewVirtualDevice("1", ciph, inCh, mock.NewRadioParams())

	outCh := dv.OutChan()

	var wg sync.WaitGroup

	// 消费者
	go func() {
		wg.Add(1)
		defer wg.Done()
		for msg := range outCh {
			//time.Sleep(time.Second)
			log.Printf("Received message: %s", msg)
		}
	}()

	dv.Send("2", []byte("hello"))
	dv.Send("3", []byte("world"))
	dv.Send("4", []byte("!!"))
	//dv.CancelSend()

	time.Sleep(3 * time.Second)
	dv.Stop()
	wg.Wait()
}

func TestDeviceReceive(t *testing.T) {

	inCh := make(chan message.Message)
	ciph := mock.NewCipher()
	dv := NewVirtualDevice("1", ciph, inCh, mock.NewRadioParams())
	go dv.Run()

	for i := 0; i < 10; i++ {
		inCh <- message.Message{
			Payload: []byte(strconv.Itoa(i)),
		}
		time.Sleep(time.Millisecond * 100)
	}

}
