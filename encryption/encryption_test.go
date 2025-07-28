package encryption

import (
	"os"
	"testing"
)

func TestEncryption(t *testing.T) {
	os.Setenv("SCWHWAB_APP_AES_GCM_KEY", "12345678901234567890123456789012")
	t.Cleanup(func() {
		os.Unsetenv("SCWHWAB_APP_AES_GCM_KEY")
	})
	tests := []struct {
		input    string
		filename string
	}{
		{
			input:    "1234",
			filename: "test.txt",
		},
	}
	for _, test := range tests {
		err := EncryptToFile([]byte(test.input), test.filename)
		if err != nil {
			t.Fatalf("failed to encrypt: %v", err)
		}
		text, err := DecryptFromFile(test.filename)
		if err != nil {
			t.Fatalf("failed to decrypt: %v", err)
		}
		if string(text) != test.input {
			t.Fatalf("expected %s, got %s", test.input, string(text))
		}
		os.Remove(test.filename)
	}
}
