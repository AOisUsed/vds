package address

// Address 地址抽象
type Address interface {
	Send(data []byte) error
}
