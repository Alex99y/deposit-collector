package crypto

import (
	hex "encoding/hex"
	fmt "fmt"

	utils "deposit-collector/pkg/utils"

	btcutil "github.com/btcsuite/btcd/btcutil"
	hdkeychain "github.com/btcsuite/btcd/btcutil/hdkeychain"
	chaincfg "github.com/btcsuite/btcd/chaincfg"
)

/**
* The source of this code is from the following repository:
* https://github.com/X-Vlad/go-hdwallet/blob/main/networks/bitcoin.go
* I have modified the code to fit my needs.
* All credits go to the author of the repository.
**/

type BitcoinWallet struct {
	Address    string
	PrivateKey string
	PublicKey  string
	Path       string
	WIF        string
}

func GenerateBitcoinWallet(
	seed []byte,
	coinType uint32,
	accountIndex uint32,
	changeIndex uint32,
	index uint32,
) (*BitcoinWallet, error) {
	var params *chaincfg.Params
	switch coinType {
	case 0:
		params = &chaincfg.MainNetParams
	case 1:
		params = &chaincfg.TestNet3Params
	default:
		return nil, utils.NewError("invalid coin type")
	}
	if coinType == 84 {
		return generateNativeSegwitWallet(
			seed, params, coinType, accountIndex, changeIndex, index,
		)
	}
	return nil, utils.NewError("bitcoin coin type not supported")
}

func generateNativeSegwitWallet(
	seed []byte,
	params *chaincfg.Params,
	coinType uint32,
	accountIndex uint32,
	changeIndex uint32,
	index uint32,
) (*BitcoinWallet, error) {
	key, err := deriveKey(
		seed, params, 84, coinType, accountIndex, changeIndex, index,
	)
	if err != nil {
		return nil, err
	}

	privKey, err := key.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	pubKey, err := key.ECPubKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Generate P2WPKH address
	witnessProg := btcutil.Hash160(pubKey.SerializeCompressed())
	address, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	// WIF
	wif, err := btcutil.NewWIF(privKey, params, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create WIF: %w", err)
	}
	pathStruct := NewBIP44(84, coinType, accountIndex, changeIndex, index)
	path := pathStruct.GeneratePath()

	return &BitcoinWallet{
		Address:    address.String(),
		PrivateKey: hex.EncodeToString(privKey.Serialize()),
		PublicKey:  hex.EncodeToString(pubKey.SerializeCompressed()),
		Path:       path,
		WIF:        wif.String(),
	}, nil
}

// deriveKey derives a key at the given path
func deriveKey(
	seed []byte,
	params *chaincfg.Params,
	purpose, coinType uint32,
	accountIndex uint32,
	changeIndex uint32,
	index uint32,
) (*hdkeychain.ExtendedKey, error) {
	masterKey, err := hdkeychain.NewMaster(seed, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create master key: %w", err)
	}

	// m/purpose'
	purposeKey, err := masterKey.Derive(hdkeychain.HardenedKeyStart + purpose)
	if err != nil {
		return nil, fmt.Errorf("failed to derive purpose: %w", err)
	}

	// m/purpose'/coin'
	coinKey, err := purposeKey.Derive(hdkeychain.HardenedKeyStart + coinType)
	if err != nil {
		return nil, fmt.Errorf("failed to derive coin type: %w", err)
	}

	// m/purpose'/coin'/0'
	accountKey, err := coinKey.Derive(hdkeychain.HardenedKeyStart + accountIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to derive account: %w", err)
	}

	// m/purpose'/coin'/'account'/0
	changeKey, err := accountKey.Derive(changeIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to derive change: %w", err)
	}

	// m/purpose'/coin'/'account'/change/0
	addressKey, err := changeKey.Derive(index)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address: %w", err)
	}

	return addressKey, nil
}
