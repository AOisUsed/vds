package mock

type Cipher struct {
}

func NewMockCipher() *Cipher {
	return &Cipher{}
}

func (m Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	return plaintext, nil
}

func (m Cipher) Decrypt(ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}
