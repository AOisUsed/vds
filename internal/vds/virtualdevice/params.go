package virtualdevice

type Params interface {
	IsCompatibleWith(other Params) bool
}
