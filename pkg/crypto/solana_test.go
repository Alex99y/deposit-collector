package crypto

import (
	"bytes"
	"testing"
)

// Solana BIP44 path: m/44'/501'/0'/0/0 (501 is Solana's coin type)
const solanaPath = "m/44'/501'/0'/0/0"

func TestGenerateSolanaWallet_ValidInput(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, 32)

	address := "CQvWgQkJqDSQjq5b4JZnH5DKveCXgkDzwBfoj8yXdbeG"
	privateKey := "74cd471c31e78bdb232bb04fe68e628c093ab20efdfeb751244924d346230079"
	publicKey := "CQvWgQkJqDSQjq5b4JZnH5DKveCXgkDzwBfoj8yXdbeG"

	wallet, err := GenerateSolanaWallet(seed, solanaPath)
	if err != nil {
		t.Fatalf("GenerateSolanaWallet() unexpected error: %v", err)
	}
	if wallet == nil {
		t.Fatal("GenerateSolanaWallet() returned nil wallet")
	}
	if wallet.Address != address {
		t.Errorf("Address = %q, want %q", wallet.Address, address)
	}
	if wallet.PrivateKey != privateKey {
		t.Errorf("PrivateKey = %q, want %q", wallet.PrivateKey, privateKey)
	}
	if wallet.PublicKey != publicKey {
		t.Errorf("PublicKey = %q, want %q", wallet.PublicKey, publicKey)
	}
	if wallet.Path != solanaPath {
		t.Errorf("Path = %q, want %q", wallet.Path, solanaPath)
	}
	// Solana address is base58-encoded 32-byte pubkey (typically 43-44 chars)
	if len(wallet.Address) < 32 || len(wallet.Address) > 44 {
		t.Errorf("Address length = %d, expected base58 pubkey length 32-44", len(wallet.Address))
	}
	// Private key is hex (32 bytes = 64 hex chars)
	if len(wallet.PrivateKey) != 64 {
		t.Errorf("PrivateKey hex length = %d, want 64", len(wallet.PrivateKey))
	}

}

func TestGenerateSolanaWallet_InvalidPath(t *testing.T) {
	seed := make([]byte, 32)

	tests := []struct {
		name string
		path string
	}{
		{"empty path", ""},
		{"wrong segment count", "m/44'/501'"},
		{"does not start with m", "x/44'/501'/0'/0/0"},
		{"invalid purpose", "m/44/501'/0'/0/0"},
		{"invalid coin type", "m/44'/xx'/0'/0/0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := GenerateSolanaWallet(seed, tt.path)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if wallet != nil {
				t.Error("expected nil wallet on error")
			}
		})
	}
}
