// Package cipher 密码机包含了加密和解密的功能
package cipher

// Cipher 接口定义了加密和解密的方法
type Cipher interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

// Plain 明文密码机(无加密解密)
type Plain struct {
}

func NewPlain() *Plain {
	return &Plain{}
}

func (p Plain) Encrypt(plaintext []byte) ([]byte, error) {
	return plaintext, nil
}

func (p Plain) Decrypt(ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}
