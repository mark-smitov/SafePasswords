package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	vaultFile = "passwords/vault.enc"
	metaFile  = "passwords/vault.meta"
)

func vaultExists() bool {
	_, err := os.Stat(vaultFile)
	return err == nil
}

func metaExists() bool {
	_, err := os.Stat(metaFile)
	return err == nil
}

func saveVault(vault *Vault, password string) error {
	plain, err := json.MarshalIndent(vault, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal vault: %w", err)
	}

	meta := &VaultMeta{Version: vaultVersion}
	if metaExists() {
		oldMeta, err := loadMeta()
		if err == nil {
			meta = oldMeta
		}
	}

	
	meta.Nonce = nil
	meta.Salt = nil

	cipher, err := encryptVault(plain, password, meta)
	if err != nil {
		return fmt.Errorf("encrypt vault: %w", err)
	}

	if err := os.WriteFile(vaultFile, cipher, 0600); err != nil {
		return fmt.Errorf("write vault: %w", err)
	}
	if err := saveMeta(meta); err != nil {
		return fmt.Errorf("write meta: %w", err)
	}
	return nil
}

func loadVault(password string) (*Vault, error) {
	cipher, err := os.ReadFile(vaultFile)
	if err != nil {
		return nil, fmt.Errorf("read vault: %w", err)
	}
	meta, err := loadMeta()
	if err != nil {
		return nil, fmt.Errorf("read meta: %w", err)
	}

	plain, err := decryptVault(cipher, password, meta)
	if err != nil {
		return nil, err
	}

	var vault Vault
	if err := json.Unmarshal(plain, &vault); err != nil {
		return nil, fmt.Errorf("unmarshal vault: %w", err)
	}
	return &vault, nil
}

func saveMeta(meta *VaultMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaFile, data, 0644)
}

func loadMeta() (*VaultMeta, error) {
	data, err := os.ReadFile(metaFile)
	if err != nil {
		return nil, err
	}
	var meta VaultMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func ensurePasswordsDir() error {
	return os.MkdirAll(filepath.Dir(vaultFile), 0700)
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
