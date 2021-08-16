package provider

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

type Payload struct {
	Data string `json:"data"`
	HMAC string `json:"hmac"`
	IV   string `json:"iv"`
}

func (p *Payload) Marshal() string {
	r, _ := json.Marshal(p)
	return string(r)
}

func Encrypt(data, key []byte) (*Payload, error) {
	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	data = aesEncrypt(data, key, iv)

	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	h.Write(iv)
	hmac := h.Sum(nil)

	return &Payload{
		Data: hex.EncodeToString(data),
		HMAC: hex.EncodeToString(hmac),
		IV:   hex.EncodeToString(iv),
	}, nil
}

func Decrypt(paylod *Payload, key []byte) ([]byte, error) {
	data, err := hex.DecodeString(paylod.Data)
	if err != nil {
		return nil, err
	}
	iv, err := hex.DecodeString(paylod.IV)
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha256.New, key)
	h.Write(data)
	h.Write(iv)
	hmac := h.Sum(nil)
	if hex.EncodeToString(hmac) != paylod.HMAC {
		return nil, errors.New("HMAC not match")
	}

	data = aesDecrypt(data, key, iv)
	return data, nil
}

func aesEncrypt(data, key, iv []byte) []byte {
	blk, _ := aes.NewCipher(key)
	mode := cipher.NewCBCEncrypter(blk, iv)
	size := blk.BlockSize()
	n := len(data) % size // pkcs7
	if n == 0 {
		n = size
	} else {
		n = size - n
	}
	d := make([]byte, len(data)+n)
	copy(d, data)
	for i := 0; i < n; i++ {
		d[len(data)+i] = byte(n)
	}
	data = d
	dst := make([]byte, len(data))
	mode.CryptBlocks(dst, data)
	return dst
}

func aesDecrypt(data, key, iv []byte) []byte {
	blk, _ := aes.NewCipher(key)
	mode := cipher.NewCBCDecrypter(blk, iv)
	dst := make([]byte, len(data))
	mode.CryptBlocks(dst, data)
	n := dst[len(dst)-1]
	return dst[:len(dst)-int(n)]
}
