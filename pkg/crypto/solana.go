package crypto

/**
* The source of this code is from the following repository:
* https://github.com/X-Vlad/go-hdwallet/blob/main/networks/solana.go
* I have modified the code to fit my needs.
* All credits go to the author of the repository.
**/

import (
	ed25519 "crypto/ed25519"
	hmac "crypto/hmac"
	sha512 "crypto/sha512"
	binary "encoding/binary"
	hex "encoding/hex"

	utils "deposit-collector/pkg/utils"

	base58 "github.com/btcsuite/btcd/btcutil/base58"
)

type SolanaWallet struct {
	Address    string
	PrivateKey string
	PublicKey  string
	Path       string
}

func deriveSolanaKey(seed []byte, path []uint32) ([]byte, error) {
	mac := hmac.New(sha512.New, []byte("ed25519 seed"))
	mac.Write(seed)
	I := mac.Sum(nil)

	key := I[:32]
	chainCode := I[32:]

	for _, childIndex := range path {
		// All Ed25519 derivations must be hardened
		hardenedIndex := childIndex + 0x80000000

		data := make([]byte, 37)
		data[0] = 0x00
		copy(data[1:33], key)
		binary.BigEndian.PutUint32(data[33:], hardenedIndex)

		mac = hmac.New(sha512.New, chainCode)
		mac.Write(data)
		I = mac.Sum(nil)

		key = I[:32]
		chainCode = I[32:]
	}

	return key, nil
}

// In Solana, the path used is a four level path:
// purpose / coin_type / account / change. I.e: m/44'/501'/{account_index}'/0'
// It is used by Phantom, Solflare, Ledger compatible. As we are not
// interested to maintain compatibility with other wallets,
// we will use the full path.
func GenerateSolanaWallet(seed []byte, path string) (*SolanaWallet, error) {
	pathStruct, err := validateBIP44Path(path)
	if err != nil {
		return nil, utils.NewError("invalid BIP44 path: " + err.Error())
	}
	// 4-level path: purpose / coin_type / account / change
	privateKey, err := deriveSolanaKey(
		seed,
		[]uint32{
			pathStruct.Purpose,
			pathStruct.CoinType,
			pathStruct.Account,
			pathStruct.Change,
			pathStruct.Index,
		},
	)
	if err != nil {
		return nil, err
	}

	edPrivKey := ed25519.NewKeyFromSeed(privateKey)
	edPubKey := edPrivKey.Public().(ed25519.PublicKey)

	address := base58.Encode(edPubKey)

	return &SolanaWallet{
		Address:    address,
		PrivateKey: hex.EncodeToString(privateKey),
		PublicKey:  base58.Encode(edPubKey),
		Path:       path,
	}, nil
}
