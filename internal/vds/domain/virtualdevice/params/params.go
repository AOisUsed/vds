package params

type Params interface {
	IsCompatibleWith(other Params) bool
}

// Empty 无类型，内容 参数
type Empty struct{}

func NewEmpty() *Empty {
	return &Empty{}
}

func (e Empty) IsCompatibleWith(other Params) bool {
	_, ok := other.(*Empty)
	if !ok {
		return false
	}
	return true
}
