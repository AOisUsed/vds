package mock

import "virturalDevice/pkg/vds/radiomodel"

type RadioParams struct {
	Mode       int
	IsOn       bool
	CryptoMode int
}

func (p *RadioParams) IsCompatibleWith(other radiomodel.Params) bool {
	otherP, ok := other.(*RadioParams)
	if !ok {
		return false
	}
	return p.Mode == otherP.Mode &&
		p.IsOn == otherP.IsOn &&
		p.CryptoMode == otherP.CryptoMode
}
