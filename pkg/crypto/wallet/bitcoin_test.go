package wallet

import (
	"bytes"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

func TestGenerateBitcoinWallet_ValidInput_Mainnet(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, 32)
	// coinType 0 = mainnet, BIP84 Native SegWit
	wallet, err := GenerateBitcoinWallet(
		seed, false, PurposeBTCNativeSegwit, 0, 0, 0,
	)
	if err != nil {
		t.Fatalf("GenerateBitcoinWallet() unexpected error: %v", err)
	}
	if wallet == nil {
		t.Fatal("GenerateBitcoinWallet() returned nil wallet")
	}
	assertBitcoinWalletFields(t, wallet)
	if !strings.HasPrefix(wallet.Address, "bc1") {
		t.Errorf(
			"mainnet SegWit address should start with bc1, got %q",
			wallet.Address,
		)
	}
	if wallet.Path != "m/84'/0'/0'/0/0" {
		t.Errorf("Path = %q, want m/84'/0'/0'/0/0", wallet.Path)
	}
}

func TestGenerateBitcoinWallet_ValidInput_Testnet(t *testing.T) {
	seed := bytes.Repeat([]byte{0x42}, 32)
	// coinType 1 = testnet
	wallet, err := GenerateBitcoinWallet(
		seed, true, PurposeBTCNativeSegwit, 0, 0, 0,
	)
	if err != nil {
		t.Fatalf("GenerateBitcoinWallet() unexpected error: %v", err)
	}
	if wallet == nil {
		t.Fatal("GenerateBitcoinWallet() returned nil wallet")
	}
	assertBitcoinWalletFields(t, wallet)
	if !strings.HasPrefix(wallet.Address, "tb1") {
		t.Errorf(
			"testnet SegWit address should start with tb1, got %q",
			wallet.Address,
		)
	}
	if wallet.Path != "m/84'/1'/0'/0/0" {
		t.Errorf("Path = %q, want m/84'/1'/0'/0/0", wallet.Path)
	}
}

func TestGenerateBitcoinWallet_InvalidCoinType(t *testing.T) {
	seed := make([]byte, 32)
	_, err := GenerateBitcoinWallet(seed, false, 99, 0, 0, 0)
	if err == nil {
		t.Fatal("expected error for invalid bitcoin purpose not supported, got nil")
	}
	if !strings.Contains(err.Error(), "bitcoin purpose not supported") {
		t.Errorf("error should mention invalid coin type: %v", err)
	}
}

func TestGenerateBitcoinWallet_Deterministic(t *testing.T) {
	seed := bytes.Repeat([]byte{0xab}, 32)

	w1, err := GenerateBitcoinWallet(seed, false, PurposeBTCNativeSegwit, 0, 0, 0)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	w2, err := GenerateBitcoinWallet(seed, false, PurposeBTCNativeSegwit, 0, 0, 0)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if w1.Address != w2.Address {
		t.Errorf("Address not deterministic: %q vs %q", w1.Address, w2.Address)
	}
	if w1.PrivateKey != w2.PrivateKey {
		t.Error("PrivateKey not deterministic")
	}
	if w1.WIF != w2.WIF {
		t.Error("WIF not deterministic")
	}
}

func TestGenerateBitcoinWallet_DifferentIndicesDifferentAddrs(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}

	w1, err := GenerateBitcoinWallet(seed, false, PurposeBTCNativeSegwit, 0, 0, 0)
	if err != nil {
		t.Fatalf("index 0: %v", err)
	}
	w2, err := GenerateBitcoinWallet(seed, false, PurposeBTCNativeSegwit, 0, 0, 1)
	if err != nil {
		t.Fatalf("index 1: %v", err)
	}
	if w1.Address == w2.Address {
		t.Error("different index should produce different address")
	}
}

func TestGenerateBitcoinWallet_InvalidSeed(t *testing.T) {
	shortSeed := []byte{0x01}
	_, err := GenerateBitcoinWallet(
		shortSeed, false, PurposeBTCNativeSegwit, 0, 0, 0,
	)
	if err == nil {
		t.Fatal("expected error for short seed, got nil")
	}
}

func TestGenerateNativeSegwitWallet_ValidInput(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, 32)
	params := &chaincfg.MainNetParams

	wallet, err := generateNativeSegwitWallet(seed, params, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("GenerateNativeSegwitWallet() unexpected error: %v", err)
	}
	if wallet == nil {
		t.Fatal("GenerateNativeSegwitWallet() returned nil wallet")
	}
	assertBitcoinWalletFields(t, wallet)
	if !strings.HasPrefix(wallet.Address, "bc1") {
		t.Errorf("address should start with bc1, got %q", wallet.Address)
	}
	// WIF for mainnet compressed typically starts with L or K
	if wallet.WIF == "" {
		t.Error("expected non-empty WIF")
	}
}

func TestDeriveKey_InvalidSeed(t *testing.T) {
	shortSeed := []byte{0x01}
	_, err := deriveKey(
		shortSeed,
		&chaincfg.MainNetParams,
		84, 0, 0, 0, 0,
	)
	if err == nil {
		t.Fatal("expected error for short seed, got nil")
	}
}

func TestDeriveKey_Deterministic(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, 32)
	params := &chaincfg.MainNetParams

	k1, err := deriveKey(seed, params, 84, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("first deriveKey: %v", err)
	}
	k2, err := deriveKey(seed, params, 84, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("second deriveKey: %v", err)
	}
	if k1.String() != k2.String() {
		t.Error("deriveKey not deterministic")
	}
}

func TestNewBIP44_GeneratePath(t *testing.T) {
	p := NewBIP44(84, 0, 0, 0, 0)
	path := p.GeneratePath()
	if path != "m/84'/0'/0'/0/0" {
		t.Errorf("GeneratePath() = %q, want m/84'/0'/0'/0/0", path)
	}
}

func assertBitcoinWalletFields(t *testing.T, w *BitcoinWallet) {
	t.Helper()
	if w.Address == "" {
		t.Error("expected non-empty Address")
	}
	if w.PrivateKey == "" {
		t.Error("expected non-empty PrivateKey")
	}
	if w.PublicKey == "" {
		t.Error("expected non-empty PublicKey")
	}
	if w.Path == "" {
		t.Error("expected non-empty Path")
	}
	if w.WIF == "" {
		t.Error("expected non-empty WIF")
	}
	// Private key is 32 bytes = 64 hex chars
	if len(w.PrivateKey) != 64 {
		t.Errorf("PrivateKey hex length = %d, want 64", len(w.PrivateKey))
	}
	// Compressed public key is 33 bytes = 66 hex chars
	if len(w.PublicKey) != 66 {
		t.Errorf("PublicKey hex length = %d, want 66", len(w.PublicKey))
	}
}
