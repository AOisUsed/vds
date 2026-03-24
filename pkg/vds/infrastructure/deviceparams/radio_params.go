package deviceparams

import (
	"virturalDevice/pkg/vds/domain/virtualdevice/params"
)

type RadioParams struct {
	Mode       int
	IsOn       bool
	CryptoMode int
	TypeTag    string `json:"type"`
}

type Option func(*RadioParams)

func WithMode(mode int) Option {
	return func(p *RadioParams) {
		p.Mode = mode
	}
}

func WithIsOn(isOn bool) Option {
	return func(p *RadioParams) {
		p.IsOn = isOn
	}
}
func WithCryptoMode(cryptoMode int) Option {
	return func(p *RadioParams) {
		p.CryptoMode = cryptoMode
	}
}

func NewRadioParams(ops ...Option) *RadioParams {

	var p = RadioParams{
		Mode:       0,
		IsOn:       false,
		CryptoMode: 0,
		TypeTag:    "RadioParams",
	}

	for _, op := range ops {
		op(&p)
	}
	return &p
}

func (p *RadioParams) IsCompatibleWith(other params.Params) bool {
	if p.Type() != other.Type() {
		return false
	}
	otherRaPa := other.(*RadioParams)
	return p.Mode == otherRaPa.Mode &&
		p.IsOn == otherRaPa.IsOn &&
		p.CryptoMode == otherRaPa.CryptoMode
}

func (p *RadioParams) Type() string {
	return p.TypeTag
}
