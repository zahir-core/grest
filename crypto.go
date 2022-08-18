package grest

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"

	"golang.org/x/crypto/hkdf"
)

var (
	CryptoKey  = "key"
	CryptoSalt = "salt"
	CryptoInfo = "info"
)

type Crypto struct {
	Key  string
	Salt string
	Info string
}

func NewCrypto(keys ...string) Crypto {
	c := Crypto{
		Key:  CryptoKey,
		Salt: CryptoSalt,
		Info: CryptoInfo,
	}
	if len(keys) > 0 {
		c.Key = keys[0]
	}
	if len(keys) > 1 {
		c.Salt = keys[1]
	}
	if len(keys) > 2 {
		c.Info = keys[2]
	}
	return c
}

func (c Crypto) Encrypt(text string) (string, error) {
	key, err := c.GenerateKey()
	if err != nil {
		return "", NewError(http.StatusInternalServerError, err.Error())
	}

	// CBC mode works on blocks so plaintexts may need to be padded to the
	// next whole block. For an example of such padding, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. Here we'll
	// assume that the plaintext is already of the correct length.
	plaintext := []byte(text)
	plaintext = c.PKCS5Padding(plaintext, aes.BlockSize)
	if len(plaintext)%aes.BlockSize != 0 {
		return "", NewError(http.StatusInternalServerError, "plaintext is not a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", NewError(http.StatusInternalServerError, err.Error())
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	// ciphertext length = iv + ciphertext
	// iv length = 1 block aesBlock
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", NewError(http.StatusInternalServerError, err.Error())
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c Crypto) Decrypt(text string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", NewError(http.StatusInternalServerError, err.Error())
	}
	if len(ciphertext) < aes.BlockSize {
		return "", NewError(http.StatusInternalServerError, "ciphertext too short")
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", NewError(http.StatusInternalServerError, "ciphertext is not a multiple of the block size, please use the correct key")
	}

	key, err := c.GenerateKey()
	if err != nil {
		return "", NewError(http.StatusInternalServerError, err.Error())
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", NewError(http.StatusInternalServerError, err.Error())
	}
	// extract iv content
	iv := ciphertext[:aes.BlockSize]
	// extract ciphertext content
	ct := ciphertext[aes.BlockSize:]

	plaintext := make([]byte, len(ciphertext)-aes.BlockSize)
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ct)
	plaintext, err = c.PKCS5Unpadding(plaintext)
	return string(plaintext), err
}

func (c Crypto) GenerateKey() ([]byte, error) {
	if len(c.Key) == 0 {
		return nil, NewError(http.StatusInternalServerError, "key cannot be empty")
	}

	key := make([]byte, 32)
	h := hkdf.New(sha256.New, []byte(c.Key), []byte(c.Salt), []byte(c.Info))
	n, err := h.Read(key)
	if err != nil {
		return nil, NewError(http.StatusInternalServerError, err.Error())
	}

	if n < 32 {
		return nil, NewError(http.StatusInternalServerError, "key too short")
	}
	return key, nil
}

func (Crypto) PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (Crypto) PKCS5Unpadding(encrypt []byte) ([]byte, error) {
	padding := encrypt[len(encrypt)-1]
	length := len(encrypt) - int(padding)
	if length > 0 {
		return encrypt[:length], nil
	}
	return encrypt, NewError(http.StatusInternalServerError, "ciphertext is not a multiple of the block size, please use the correct key")
}
