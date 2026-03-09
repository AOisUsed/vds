package radiomodel

type Params interface {
	IsCompatibleWith(other Params) bool
}
