package mock

import (
	"virturalDevice/internal/vds/virtualdevice/params"
)

type RadioParams struct {
	Mode       int
	IsOn       bool
	CryptoMode int
}

type Options func(*RadioParams)

func (p *RadioParams) WithMode(mode int) Options {
	return func(p *RadioParams) {
		p.Mode = mode
	}
}

func (p *RadioParams) WithIsOn(isOn bool) Options {
	return func(p *RadioParams) {
		p.IsOn = isOn
	}
}
func (p *RadioParams) WithCryptoMode(cryptoMode int) Options {
	return func(p *RadioParams) {
		p.CryptoMode = cryptoMode
	}
}

func NewRadioParams(ops ...Options) *RadioParams {

	var p = RadioParams{
		Mode:       0,
		IsOn:       false,
		CryptoMode: 0,
	}

	for _, op := range ops {
		op(&p)
	}
	return &p
}

func (p *RadioParams) IsCompatibleWith(other params.Params) bool {
	otherRaPa, ok := other.(*RadioParams)
	if !ok {
		return false
	}
	return p.Mode == otherRaPa.Mode &&
		p.IsOn == otherRaPa.IsOn &&
		p.CryptoMode == otherRaPa.CryptoMode
}
