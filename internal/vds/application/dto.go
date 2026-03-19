package application

type StartVDSRequest struct{}

type StartVDSResponse struct {
	Success bool
	Message string
}

type StopVDSRequest struct{}

type StopVDSResponse struct {
	Success bool
}

type StartVDRequest struct {
	id string
}

type StartVDResponse struct {
	Success bool
	Message string
}

type StopVDRequest struct {
	id string
}
type StopVDResponse struct {
	Success bool
	Message string
}

type SendMessageRequest struct {
	Src     string
	Dst     string
	Payload []byte
}

type SendMessageResponse struct {
	Success bool
	Message string
}

type CancelMessageRequest struct{}

// CancelMessageResponse  // 可能应该加入操作是否成功的回复？可能无法确认是否成功
type CancelMessageResponse struct{}

type SubscribeDeviceMessageRequest struct {
	DeviceID     string
	SubscriberID string
}

type SubscribeDeviceMessageResponse struct {
	Success bool
	Message string
}
