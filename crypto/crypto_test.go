package crypto

import (
	"testing"
)

func newEncryptDecryptTestData() []string {
	return []string{
		"Lorem ipsum dolor sit amet,",
		"consectetur adipisicing elit,",
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam,",
		"quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. ",
		"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. ",
		"Excepteur sint occaecat cupidatat non proident,",
		"sunt in culpa qui officia deserunt mollit anim id est laborum.",
	}
}

func TestEncryptDecrypt(t *testing.T) {
	Configure("972ec8dd995743d981417981ac2f30db")
	data := newEncryptDecryptTestData()
	for _, plaintext := range data {
		encrypted, err := Encrypt(plaintext)
		if err != nil {
			t.Errorf("Test encrypt : Error occurred [%v]", err)
		}
		if encrypted == plaintext {
			t.Errorf("Test encrypt : encrypted must not be equal to [%v]", plaintext)
		}
		decrypted, err := Decrypt(encrypted)
		if err != nil {
			t.Errorf("Test decrypt : Error occurred [%v]", err)
		}
		if decrypted != plaintext {
			t.Errorf("Test decrypt : expected decrypted [%v], got [%v]", plaintext, decrypted)
		}
	}
}

func BenchmarkEncryptDecrypt(b *testing.B) {
	data := newEncryptDecryptTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			encrypted, err := Encrypt(d)
			if err == nil {
				Decrypt(encrypted)
			}
		}
	}
}
