package wallet

import (
	bytes "bytes"
	hex "encoding/hex"
	errors "errors"
	fmt "fmt"

	btc_utils "deposit-collector/pkg/crypto/btc"

	btcutil "github.com/btcsuite/btcd/btcutil"
	hdkeychain "github.com/btcsuite/btcd/btcutil/hdkeychain"
	chaincfg "github.com/btcsuite/btcd/chaincfg"
	txscript "github.com/btcsuite/btcd/txscript"
	wire "github.com/btcsuite/btcd/wire"
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

type BitcoinTransactionInput struct {
	Index      int
	AmountSats int64
}

func (b *BitcoinWallet) GetAddress() string {
	return b.Address
}

func (b *BitcoinWallet) SignMessage(message string) ([]byte, error) {
	return nil, nil
}

func (b *BitcoinWallet) SignTransactionInputs(
	tx wire.MsgTx,
	network btc_utils.NETWORK,
	inputs []BitcoinTransactionInput,
) (*wire.MsgTx, error) {

	netParams := btc_utils.GetNetParamsByNetwork(network)
	if netParams == nil {
		return nil, errors.New("invalid bitcoin network")
	}

	wif, err := btcutil.DecodeWIF(b.WIF)
	if err != nil {
		return nil, err
	}

	address, err := btcutil.DecodeAddress(b.Address, netParams)
	if err != nil {
		return nil, err
	}
	pkScript, err := txscript.PayToAddrScript(address)
	if err != nil {
		return nil, err
	}

	prevOutFetcher := txscript.NewMultiPrevOutFetcher(nil)
	for _, input := range inputs {
		if input.Index < 0 || input.Index >= len(tx.TxIn) {
			return nil, errors.New("invalid tx input index")
		}
		prevOut := tx.TxIn[input.Index].PreviousOutPoint
		prevOutFetcher.AddPrevOut(
			prevOut,
			&wire.TxOut{
				Value:    input.AmountSats,
				PkScript: pkScript,
			},
		)
	}

	sigHashes := txscript.NewTxSigHashes(&tx, prevOutFetcher)
	for _, input := range inputs {
		witness, err := txscript.WitnessSignature(
			&tx,
			sigHashes,
			input.Index,
			input.AmountSats,
			pkScript,
			txscript.SigHashAll,
			wif.PrivKey,
			true,
		)
		if err != nil {
			return nil, err
		}
		tx.TxIn[input.Index].Witness = witness
	}

	var signedTxBuffer bytes.Buffer
	if err := tx.Serialize(&signedTxBuffer); err != nil {
		return nil, err
	}
	return &tx, nil
}

func GenerateBitcoinWallet(
	seed []byte,
	network btc_utils.NETWORK,
	purpose uint32,
	accountIndex uint32,
	changeIndex uint32,
	index uint32,
) (*BitcoinWallet, error) {
	params := btc_utils.GetNetParamsByNetwork(network)
	coinType := params.HDCoinType
	if purpose == PurposeBTCNativeSegwit {
		return generateNativeSegwitWallet(
			seed, params, coinType, accountIndex, changeIndex, index,
		)
	}
	return nil, errors.New("bitcoin purpose not supported")
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
		seed, params, PurposeBTCNativeSegwit,
		coinType, accountIndex, changeIndex, index,
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
