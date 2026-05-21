package internal

import (
	"time"
)


type Entry struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	URL       string    `json:"url,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


type Vault struct {
	Version  int       `json:"version"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
	Entries  []Entry   `json:"entries"`
}


type VaultMeta struct {
	Version      int    `json:"version"`
	Salt         []byte `json:"salt"`
	Nonce        []byte `json:"nonce"`
	ArgonMemory  uint32 `json:"argon_memory"`
	ArgonTime    uint32 `json:"argon_time"`
	ArgonThreads uint8  `json:"argon_threads"`
	KeyLen       uint32 `json:"key_len"`
}

const vaultVersion = 1
