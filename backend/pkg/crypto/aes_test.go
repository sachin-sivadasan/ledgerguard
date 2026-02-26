package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 bytes for AES-256
	plaintext := []byte("my-secret-access-token")

	encryptor, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext should not equal plaintext")
	}

	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted text does not match original: got %s, want %s", decrypted, plaintext)
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	plaintext := []byte("my-secret-access-token")

	encryptor, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	ciphertext1, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	ciphertext2, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	// Due to random nonce, same plaintext should produce different ciphertext
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("encrypting same plaintext twice should produce different ciphertext")
	}
}

func TestInvalidKeyLength(t *testing.T) {
	key := []byte("short-key")

	_, err := NewAESEncryptor(key)
	if err == nil {
		t.Error("expected error for invalid key length")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key := []byte("12345678901234567890123456789012")

	encryptor, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	// Too short ciphertext
	_, err = encryptor.Decrypt([]byte("short"))
	if err == nil {
		t.Error("expected error for invalid ciphertext")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := []byte("12345678901234567890123456789012")
	key2 := []byte("abcdefghijklmnopqrstuvwxyz123456")
	plaintext := []byte("my-secret-access-token")

	encryptor1, _ := NewAESEncryptor(key1)
	encryptor2, _ := NewAESEncryptor(key2)

	ciphertext, err := encryptor1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	_, err = encryptor2.Decrypt(ciphertext)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
}
