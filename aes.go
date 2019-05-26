package locknut

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
)

// Cryptor used for aes ops
type Cryptor struct {
	plain  []byte
	Cipher []byte `json:"cipher"`
}

// NewCryptor construct new aes object
func NewCryptor(plain []byte) Cryptor {
	return Cryptor{plain: plain}
}

// PackToBytes json marshal payload
func (c *Cryptor) PackToBytes() ([]byte, error) {
	p, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// PayloadFromBytes json unmarshal payload
func (c *Cryptor) PayloadFromBytes(bpayload []byte) error {
	err := json.Unmarshal(bpayload, &c)
	return err
}

// Encrypt abstracted encryption scheme
func (c *Cryptor) Encrypt(key []byte) error {
	a, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(a)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	c.Cipher, err = gcm.Seal(nonce, nonce, c.plain, nil), nil
	return err
}

// EncryptWithNewKey abstracted encryption scheme for capsule
func (c *Cryptor) EncryptWithNewKey() ([]byte, error) {
	key, err := GetRandKey()
	if err != nil {
		return nil, err
	}
	c.Encrypt(key)
	return key, nil
}

// Decrypt get bytes from cipher
func (c *Cryptor) Decrypt(key []byte) ([]byte, error) {
	a, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(a)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(c.Cipher) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := c.Cipher[:nonceSize], c.Cipher[nonceSize:]
	c.plain, err = gcm.Open(nil, nonce, ciphertext, nil)
	return c.plain, err
}

// Decrypt without reciever
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	return plain, err
}

// Encrypt recieverless encryption
func Encrypt(plain, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plain, nil), nil
}

// GetRandKey gets cryptographically secure 32B
func GetRandKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
