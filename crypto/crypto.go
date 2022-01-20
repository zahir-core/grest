package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/hkdf"
)

var (
	cryptoKey  = "key"
	cryptoSalt = "salt"
	cryptoInfo = "info"
)

var (
	ErrKeyEmpty    = errors.New("key cannot be empty")
	ErrKeyTooShort = errors.New("key too short")
	ErrCtTooShort  = errors.New("ciphertext too short")
	ErrCtUnpadded  = errors.New("ciphertext is not a multiple of the block size")
	ErrPtUnpadded  = errors.New("plaintext is not a multiple of the block size")
)

func Configure(key string, additionalKey ...string) {
	cryptoKey = key
	if len(additionalKey) > 0 {
		cryptoSalt = additionalKey[0]
	}
	if len(additionalKey) > 1 {
		cryptoInfo = additionalKey[1]
	}
}

func Encrypt(text string) (string, error) {
	key, err := generateKey()
	if err != nil {
		return "", err
	}

	// CBC mode works on blocks so plaintexts may need to be padded to the
	// next whole block. For an example of such padding, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. Here we'll
	// assume that the plaintext is already of the correct length.
	plaintext := []byte(text)
	plaintext = pkcs5{}.Padding(plaintext, aes.BlockSize)
	if len(plaintext)%aes.BlockSize != 0 {
		return "", ErrPtUnpadded
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	// ciphertext length = iv + ciphertext
	// iv length = 1 block aesBlock
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(text string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		return "", ErrCtTooShort
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", ErrCtUnpadded
	}

	key, err := generateKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	// extract iv content
	iv := ciphertext[:aes.BlockSize]
	// extract ciphertext content
	ct := ciphertext[aes.BlockSize:]

	plaintext := make([]byte, len(ciphertext)-aes.BlockSize)
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ct)
	plaintext = pkcs5{}.Unpadding(plaintext)
	return string(plaintext), nil
}

func generateKey() ([]byte, error) {
	if len(cryptoKey) == 0 {
		return nil, ErrKeyEmpty
	}

	key := make([]byte, 32)
	h := hkdf.New(sha256.New, []byte(cryptoKey), []byte(cryptoSalt), []byte(cryptoInfo))
	n, err := h.Read(key)
	if err != nil {
		return nil, err
	}

	if n < 32 {
		return nil, ErrKeyTooShort
	}
	return key, nil
}

type pkcs5 struct{}

func (p pkcs5) Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (p pkcs5) Unpadding(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
