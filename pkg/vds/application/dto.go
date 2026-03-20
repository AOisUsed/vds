package application

type StartVDRequest struct {
	id string
}

type StopVDRequest struct {
	id string
}

type SendMessageRequest struct {
	Src     string
	Dst     string
	Payload []byte
}

type CancelMessageRequest struct{}

type SubscribeDeviceMessageRequest struct {
	DeviceID     string
	SubscriberID string
}
