package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory  = 64 * 1024 
	argonTime    = 3
	argonThreads = 1
	keyLength    = 32
	saltLength   = 16
	nonceLength  = 12
)


func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, keyLength)
}


func encryptVault(plain []byte, password string, meta *VaultMeta) ([]byte, error) {
	if len(meta.Salt) == 0 {
		meta.Salt = make([]byte, saltLength)
		if _, err := io.ReadFull(rand.Reader, meta.Salt); err != nil {
			return nil, fmt.Errorf("failed to generate salt: %w", err)
		}
	}
	if len(meta.Nonce) == 0 {
		meta.Nonce = make([]byte, nonceLength)
		if _, err := io.ReadFull(rand.Reader, meta.Nonce); err != nil {
			return nil, fmt.Errorf("failed to generate nonce: %w", err)
		}
	}
	meta.ArgonMemory = argonMemory
	meta.ArgonTime = argonTime
	meta.ArgonThreads = argonThreads
	meta.KeyLen = keyLength

	key := deriveKey(password, meta.Salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes gcm: %w", err)
	}

	ciphertext := aesgcm.Seal(nil, meta.Nonce, plain, nil)
	return ciphertext, nil
}


func decryptVault(ciphertext []byte, password string, meta *VaultMeta) ([]byte, error) {
	key := deriveKey(password, meta.Salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes gcm: %w", err)
	}

	plain, err := aesgcm.Open(nil, meta.Nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong password or corrupted data): %w", err)
	}
	return plain, nil
}
