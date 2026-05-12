package btc_utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

type NETWORK string

const (
	MAINNET NETWORK = "mainnet"
	TESTNET NETWORK = "testnet"
	LOCAL   NETWORK = "local"
)

const (
	ONE_BITCOIN_IN_SATS    = 100_000_000
	MIN_FEE_PER_KB_IN_SATS = 1000
)

func BitcoinToSatoshis(amount float64) (int64, error) {
	integerAmount := amount * ONE_BITCOIN_IN_SATS
	stringAmount := strconv.FormatFloat(integerAmount, 'f', -1, 64)
	decimalPart := strings.Split(stringAmount, ".")
	if len(decimalPart) > 1 {
		return 0, errors.New("invalid amount")
	}
	return int64(amount * ONE_BITCOIN_IN_SATS), nil
}

func GetAddressFromScript(
	scriptHex string, net *chaincfg.Params,
) (string, error) {
	scriptBytes, err := hex.DecodeString(scriptHex)
	if err != nil {
		return "", fmt.Errorf("error converting %v to hex", scriptHex)
	}

	_, addrs, _, err := txscript.ExtractPkScriptAddrs(scriptBytes, net)
	if err != nil {
		return "", fmt.Errorf("error extracting public keys from script: %v", err)
	}

	if len(addrs) > 0 {
		return addrs[0].EncodeAddress(), nil
	}

	return "", errors.New(
		"invalid public key script, deosn't contains any addresses",
	)
}

func GetNetParamsByNetwork(network NETWORK) *chaincfg.Params {
	var params *chaincfg.Params
	switch network {
	case TESTNET:
		params = &chaincfg.TestNet3Params
	case MAINNET:
		params = &chaincfg.MainNetParams
	case LOCAL:
		params = &chaincfg.RegressionNetParams
	}

	return params
}

func ValidateAddress(address string, network NETWORK) bool {
	_, err := btcutil.DecodeAddress(address, GetNetParamsByNetwork(network))
	return err == nil
}

// ElectrumScriptHashForAddress returns the Electrum scripthash (sha256(scriptPubKey),
// reversed like Bitcoin block hashes) for blockchain.scripthash.* RPC calls.
func ElectrumScriptHashForAddress(address string, network NETWORK) (string, error) {
	params := GetNetParamsByNetwork(network)
	if params == nil {
		return "", errors.New("invalid bitcoin network")
	}
	addr, err := btcutil.DecodeAddress(address, params)
	if err != nil {
		return "", err
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(pkScript)
	for i, j := 0, len(h)-1; i < j; i, j = i+1, j-1 {
		h[i], h[j] = h[j], h[i]
	}
	return hex.EncodeToString(h[:]), nil
}

// BitcoinWithdrawSigner holds native segwit (P2WPKH) address and WIF for signing
// withdrawal sweeps from collector key material (WIF or 32-byte hex private key).
type BitcoinWithdrawSigner struct {
	Address string
	WIF     string
}

// BitcoinWalletFromWithdrawCollectorKey derives a P2WPKH address and WIF from
// withdraw-collector key material. It does not import pkg/crypto/wallet to avoid
// an import cycle with that package.
func BitcoinWalletFromWithdrawCollectorKey(
	privateKey string,
	network NETWORK,
) (*BitcoinWithdrawSigner, error) {
	params := GetNetParamsByNetwork(network)
	if params == nil {
		return nil, errors.New("invalid bitcoin network")
	}
	pk := strings.TrimSpace(privateKey)
	pk = strings.TrimPrefix(pk, "0x")

	if wifKey, err := btcutil.DecodeWIF(pk); err == nil {
		if !wifKey.IsForNet(params) {
			return nil, errors.New("bitcoin WIF network does not match configured network")
		}
		witnessProg := btcutil.Hash160(
			wifKey.PrivKey.PubKey().SerializeCompressed(),
		)
		addr, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, params)
		if err != nil {
			return nil, err
		}
		return &BitcoinWithdrawSigner{
			Address: addr.String(),
			WIF:     wifKey.String(),
		}, nil
	}

	keyBytes, err := hex.DecodeString(pk)
	if err != nil {
		return nil, fmt.Errorf("bitcoin private key: decode WIF or hex: %w", err)
	}
	if len(keyBytes) != 32 {
		return nil, errors.New("bitcoin hex private key must be 32 bytes")
	}
	priv, _ := btcec.PrivKeyFromBytes(keyBytes)
	wifKey, err := btcutil.NewWIF(priv, params, true)
	if err != nil {
		return nil, err
	}
	witnessProg := btcutil.Hash160(wifKey.PrivKey.PubKey().SerializeCompressed())
	addr, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, params)
	if err != nil {
		return nil, err
	}
	return &BitcoinWithdrawSigner{
		Address: addr.String(),
		WIF:     wifKey.String(),
	}, nil
}
