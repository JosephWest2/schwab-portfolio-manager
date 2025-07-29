package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"log"
	"os"
)

var EncryptedTokenFilename = "token.txt"

func EncryptToFile(data []byte, filepath string) error {
	aesKey := os.Getenv("SCWHWAB_APP_AES_GCM_KEY")
	if aesKey == "" {
		log.Fatal("SCWHWAB_APP_AES_GCM_KEY not set")
	}

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		log.Fatal(err)
	}

	aesGCM, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		log.Fatal(err)
	}

	cipherText := aesGCM.Seal(nil, nil, data, nil)
	return os.WriteFile(filepath, cipherText, 0600)
}

func DecryptFromFile(filepath string) ([]byte, error) {
	aesKey := os.Getenv("SCWHWAB_APP_AES_GCM_KEY")
	if aesKey == "" {
		log.Fatal("SCWHWAB_APP_AES_GCM_KEY not set")
	}

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		log.Fatal(err)
	}

	aesGCM, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		log.Fatal(err)
	}

	cipherText, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	plainText, err := aesGCM.Open(nil, nil, cipherText, nil)
	if err != nil {
		log.Fatal(err)
	}

	return plainText, nil
}
