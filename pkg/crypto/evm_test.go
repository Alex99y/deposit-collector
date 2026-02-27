package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateEvmWallet_ValidInput(t *testing.T) {
	// Deterministic 32-byte seed (min length for BIP32 is 16, typical is 32)
	seed := bytes.Repeat([]byte{0x01}, 32)
	expectedAddress := "0xb6a0f727D8D6F0FdE430f7328904774898d95183"
	expectedPrivateKey := "1333518687dfbcd6f9f058c42b5eed93d9205232ac19c88566eabfd00e807e5c"
	expectedPublicKey := "04656c9dd5473505712a900575d376993496e8ea6775a7c1b8d3e8811847ae98007b1e7e824b151a7fc26a9e3cebfa3fb7beb6fa46850a680161c5c4a38bddf91d"
	path := "m/44'/60'/0'/0/0"

	wallet, err := GenerateEvmWallet(seed, path)
	if err != nil {
		t.Fatalf("GenerateEvmWallet() unexpected error: %v", err)
	}
	if wallet == nil {
		t.Fatal("GenerateEvmWallet() returned nil wallet")
	}
	if wallet.Path != path {
		t.Errorf("Path = %q, want %q", wallet.Path, path)
	}
	if wallet.Address != expectedAddress {
		t.Errorf("Address = %q, want %q", wallet.Address, expectedAddress)
	}
	if wallet.PrivateKey != expectedPrivateKey {
		t.Errorf("PrivateKey = %q, want %q", wallet.PrivateKey, expectedPrivateKey)
	}
	if wallet.PublicKey != expectedPublicKey {
		t.Errorf("PublicKey = %q, want %q", wallet.PublicKey, expectedPublicKey)
	}
}

func TestGenerateEvmWallet_InvalidPath(t *testing.T) {
	seed := make([]byte, 32)

	tests := []struct {
		name string
		path string
	}{
		{"empty path", ""},
		{"wrong segment count", "m/44'/60'"},
		{"does not start with m", "x/44'/60'/0'/0/0"},
		{"invalid purpose", "m/44/60'/0'/0/0"},
		{"invalid coin type", "m/44'/xx'/0'/0/0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := GenerateEvmWallet(seed, tt.path)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if wallet != nil {
				t.Error("expected nil wallet on error")
			}
		})
	}
}

func TestGenerateEvmWallet_InvalidSeed(t *testing.T) {
	// Seed too short for BIP32 (min 16 bytes)
	shortSeed := []byte{0x01}
	path := "m/44'/60'/0'/0/0"

	wallet, err := GenerateEvmWallet(shortSeed, path)
	if err == nil {
		t.Fatal("expected error for short seed, got nil")
	}
	if wallet != nil {
		t.Error("expected nil wallet on error")
	}
}

// func TestEvmWallet_SignMessage_Success(t *testing.T) {
// 	// Create a key with go-ethereum to get a known valid wallet
// 	key, err := crypto.GenerateKey()
// 	if err != nil {
// 		t.Fatalf("GenerateKey: %v", err)
// 	}
// 	privBytes := crypto.FromECDSA(key)
// 	privHex := hex.EncodeToString(privBytes)
// 	pubBytes := crypto.FromECDSAPub(&key.PublicKey)
// 	pubHex := hex.EncodeToString(pubBytes)
// 	addr := crypto.PubkeyToAddress(key.PublicKey)

// 	wallet := &EvmWallet{
// 		Address:    addr.Hex(),
// 		PrivateKey: privHex,
// 		PublicKey:  pubHex,
// 		Path:       "m/44'/60'/0'/0/0",
// 	}

// 	msg := "hello world"
// 	sig, err := wallet.SignMessage(msg)
// 	if err != nil {
// 		t.Fatalf("SignMessage() error: %v", err)
// 	}
// 	if len(sig) == 0 {
// 		t.Error("expected non-empty signature")
// 	}
// 	// ECDSA signature is typically 65 bytes (R || S || V)
// 	if len(sig) != 65 {
// 		t.Errorf("signature length = %d, want 65", len(sig))
// 	}

// 	// Verify: recover signer from signature and compare address
// 	hash := crypto.Keccak256Hash([]byte(msg))
// 	pubKey, err := crypto.SigToPub(hash.Bytes(), sig)
// 	if err != nil {
// 		t.Fatalf("SigToPub: %v", err)
// 	}
// 	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
// 	if recoveredAddr != addr {
// 		t.Errorf("recovered address %s, want %s", recoveredAddr.Hex(), addr.Hex())
// 	}
// }

// func TestEvmWallet_SignMessage_InvalidPrivateKey(t *testing.T) {
// 	wallet := &EvmWallet{
// 		Address:    "0x1234567890123456789012345678901234567890",
// 		PrivateKey: "not-hex",
// 		PublicKey:  "04",
// 		Path:       "m/44'/60'/0'/0/0",
// 	}

// 	sig, err := wallet.SignMessage("test")
// 	if err == nil {
// 		t.Fatal("expected error for invalid private key, got nil")
// 	}
// 	if sig != nil {
// 		t.Error("expected nil signature on error")
// 	}
// }

// func TestEvmWallet_SignMessage_EmptyMessage(t *testing.T) {
// 	key, _ := crypto.GenerateKey()
// 	wallet := &EvmWallet{
// 		PrivateKey: hex.EncodeToString(crypto.FromECDSA(key)),
// 		PublicKey:  hex.EncodeToString(crypto.FromECDSAPub(&key.PublicKey)),
// 		Address:    crypto.PubkeyToAddress(key.PublicKey).Hex(),
// 	}

// 	sig, err := wallet.SignMessage("")
// 	if err != nil {
// 		t.Fatalf("SignMessage(\"\") unexpected error: %v", err)
// 	}
// 	if len(sig) != 65 {
// 		t.Errorf("signature length = %d, want 65", len(sig))
// 	}
// }
