package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	plaintext := "gREST crypto testing"

	Configure("972ec8dd995743d981417981ac2f30db")
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
	if decrypted != plaintext {
		t.Errorf("Expected decrypted [%v], got [%v]", plaintext, decrypted)
	}
}
