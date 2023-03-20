package grest

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/cristalhq/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/hkdf"
)

var (
	CryptoKey  = "wAGyTpFQX5uKV3JInABXXEdpgFkQLPTf"
	CryptoSalt = "0de0cda7d2dd4937a1c4f7ddc43c580f"
	CryptoInfo = "info"
	JWTKey     = "f4cac8b77a8d4cb5881fac72388bb226"
)

type CryptoInterface interface {
	NewHash(text string, cost ...int) (string, error)
	CompareHash(hashed, text string) error
	NewJWT(claims any) (string, error)
	ParseAndVerifyJWT(token string, claims any) error
	Encrypt(text string) (string, error)
	Decrypt(text string) (string, error)
	GenerateKey() ([]byte, error)
	PKCS5Padding(ciphertext []byte, blockSize int) []byte
	PKCS5Unpadding(encrypt []byte) ([]byte, error)
}

type Crypto struct {
	Key    string
	Salt   string
	Info   string
	JWTKey string
}

func NewCrypto(keys ...string) *Crypto {
	c := &Crypto{
		Key:    CryptoKey,
		Salt:   CryptoSalt,
		Info:   CryptoInfo,
		JWTKey: JWTKey,
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
	if len(keys) > 3 {
		c.JWTKey = keys[3]
	}
	return c
}

func (*Crypto) NewHash(text string, cost ...int) (string, error) {
	hashCost := 10
	if len(cost) > 0 {
		hashCost = cost[0]
	}
	b, err := bcrypt.GenerateFromPassword([]byte(text), hashCost)
	return string(b), err
}

func (*Crypto) CompareHash(hashed, text string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(text))
}

func (c *Crypto) NewJWT(claims any) (string, error) {
	signer, err := jwt.NewSignerHS(jwt.HS256, []byte(c.JWTKey))
	if err != nil {
		return "", err
	}
	token, err := jwt.NewBuilder(signer).Build(claims)
	if err != nil {
		return "", err
	}
	return token.String(), nil
}

func (c *Crypto) ParseAndVerifyJWT(token string, claims any) error {
	verifier, err := jwt.NewVerifierHS(jwt.HS256, []byte(c.JWTKey))
	if err != nil {
		return err
	}
	t, err := jwt.Parse([]byte(token), verifier)
	if err != nil {
		return err
	}
	return json.Unmarshal(t.Claims(), &claims)
}

func (c *Crypto) Encrypt(text string) (string, error) {
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

func (c *Crypto) Decrypt(text string) (string, error) {
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

func (c *Crypto) GenerateKey() ([]byte, error) {
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

func (*Crypto) PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (*Crypto) PKCS5Unpadding(encrypt []byte) ([]byte, error) {
	padding := encrypt[len(encrypt)-1]
	length := len(encrypt) - int(padding)
	if length > 0 {
		return encrypt[:length], nil
	}
	return encrypt, NewError(http.StatusInternalServerError, "ciphertext is not a multiple of the block size, please use the correct key")
}
