package cryptoutil

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewCipher(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr bool
	}{
		{"valid key", make([]byte, 32), false},
		{"nil key", nil, true},
		{"empty key", []byte{}, true},
		{"16 bytes", make([]byte, 16), true},
		{"31 bytes", make([]byte, 31), true},
		{"33 bytes", make([]byte, 33), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCipher(tt.key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewCipher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	c, err := NewCipher([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello")},
		{"russian", []byte("Привет, мир!")},
		{"json", []byte(`{"user":"admin"}`)},
		{"binary", []byte{0x00, 0x01, 0xff, 0x00}},
		{"1 KB", bytes.Repeat([]byte("A"), 1024)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := c.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}

			dec, err := c.Decrypt(enc)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}

			if !bytes.Equal(dec, tt.plaintext) {
				t.Fatalf("got %v, want %v", dec, tt.plaintext)
			}
		})
	}
}

func TestDecryptErrors(t *testing.T) {
	c, _ := NewCipher([]byte("0123456789abcdef0123456789abcdef"))

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{"invalid base64", "!!!", "decode base64"},
		{"empty string", "", "too short"},
		{"too short", "AAAA", "too short"},
		{"garbage", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", "decrypt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.Decrypt(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got %q, want containing %q", err, tt.wantErr)
			}
		})
	}
}
