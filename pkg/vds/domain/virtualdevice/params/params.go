package params

// Params 参数接口
// 注： 实现该接口的实际参数一定要携带一个
// TypeTag string `json:"type"`字段以便于在反序列化时知晓应该投射到的具体类型上，
// 因为json反序列化无法投射到接口上
type Params interface {
	IsCompatibleWith(other Params) bool
	Type() string
}

// Empty 无类型，内容 参数
type Empty struct {
	TypeTag string `json:"type"`
}

func NewEmpty() *Empty {
	return &Empty{
		TypeTag: "",
	}
}

func (e Empty) IsCompatibleWith(other Params) bool {
	if e.Type() != other.Type() {
		return false
	}
	return true
}

func (e Empty) Type() string {
	return e.TypeTag
}
