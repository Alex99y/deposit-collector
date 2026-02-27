package crypto

/**
* The source of this code is from the following repository:
* https://github.com/X-Vlad/go-hdwallet/blob/main/networks/ethereum.go
* I have modified the code to fit my needs.
* All credits go to the author of the repository.
**/

import (
	ecdsa "crypto/ecdsa"
	hex "encoding/hex"

	utils "deposit-collector/pkg/utils"

	hdkeychain "github.com/btcsuite/btcd/btcutil/hdkeychain"
	chaincfg "github.com/btcsuite/btcd/chaincfg"
	crypto "github.com/ethereum/go-ethereum/crypto"
)

type EvmWallet struct {
	Address    string
	PrivateKey string
	PublicKey  string
	Path       string
}

func (e *EvmWallet) SignMessage(message string) ([]byte, error) {
	messageHash := crypto.Keccak256([]byte(message))
	privateKey, err := crypto.HexToECDSA(e.PrivateKey)
	if err != nil {
		return nil, utils.NewError(
			"failed to convert private key to ECDSA: " + err.Error(),
		)
	}
	signature, err := crypto.Sign(messageHash, privateKey)
	if err != nil {
		return nil, utils.NewError("failed to sign message: " + err.Error())
	}
	return signature, nil
}

func GenerateEvmWallet(seed []byte, path string) (*EvmWallet, error) {
	pathStruct, err := validateBIP44Path(path)
	if err != nil {
		return nil, utils.NewError("invalid BIP44 path: " + err.Error())
	}

	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, utils.NewError("failed to create master key: " + err.Error())
	}

	purposeKey, err := masterKey.Derive(
		hdkeychain.HardenedKeyStart + uint32(pathStruct.Purpose),
	)
	if err != nil {
		return nil, utils.NewError("failed to derive purpose key: " + err.Error())
	}

	coinTypeKey, err := purposeKey.Derive(
		hdkeychain.HardenedKeyStart + uint32(pathStruct.CoinType),
	)
	if err != nil {
		return nil, utils.NewError("failed to derive coin type key: " + err.Error())
	}

	accountKey, err := coinTypeKey.Derive(
		hdkeychain.HardenedKeyStart + uint32(pathStruct.Account),
	)
	if err != nil {
		return nil, utils.NewError("failed to derive account key: " + err.Error())
	}

	changeKey, err := accountKey.Derive(uint32(pathStruct.Change))
	if err != nil {
		return nil, utils.NewError("failed to derive change key: " + err.Error())
	}

	indexKey, err := changeKey.Derive(uint32(pathStruct.Index))
	if err != nil {
		return nil, utils.NewError("failed to derive index key: " + err.Error())
	}

	// Get private key bytes
	privKey, err := indexKey.ECPrivKey()
	if err != nil {
		return nil, utils.NewError("failed to get private key: " + err.Error())
	}

	// Convert to ECDSA private key for Ethereum
	privateKeyBytes := privKey.Serialize()
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, utils.NewError("failed to convert to ECDSA: " + err.Error())
	}

	// Get public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, utils.NewError("failed to cast public key to ECDSA")
	}

	// Get Ethereum address
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &EvmWallet{
		Address:    address.Hex(),
		PrivateKey: hex.EncodeToString(privateKeyBytes),
		PublicKey:  hex.EncodeToString(crypto.FromECDSAPub(publicKeyECDSA)),
		Path:       path,
	}, nil
}
