package types

type VDParams interface {
	IsCompatibleWith(other VDParams) bool
}
