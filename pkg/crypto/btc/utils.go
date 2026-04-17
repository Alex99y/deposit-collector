package btc_utils

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

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
	ONE_BITCOIN_IN_SATS = 100_000_000
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
