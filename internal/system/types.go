package system

import (
	errors "errors"

	"github.com/google/uuid"
)

type ChainPlatform string

const (
	ChainPlatformEVM ChainPlatform = "EVM"
	ChainPlatformBTC ChainPlatform = "BTC"
	ChainPlatformSOL ChainPlatform = "SOL"
)

func ValidateChainPlatform(chainPlatform string) error {
	switch chainPlatform {
	case string(ChainPlatformEVM):
		return nil
	case string(ChainPlatformBTC):
		return nil
	case string(ChainPlatformSOL):
		return nil
	default:
		return errors.New("invalid chain platform")
	}
}

type NewSupportedChainRequest struct {
	Network       string
	ChainPlatform ChainPlatform
	BIP44ID       int
}

type SupportedChain struct {
	NewSupportedChainRequest
	ID uuid.UUID
}

/*
*
Network is the network of the chain.
Example: Ethereum
*/
type NewTokenAddressRequest struct {
	UnitName   string
	UnitSymbol string
	Address    string
	Network    string
	Decimals   int
}

type TokenAddress struct {
	NewTokenAddressRequest
	ID    uuid.UUID
	Chain SupportedChain
}

type GetTokenAddressesRequest struct {
	Chain      *string
	Address    *string
	UnitSymbol *string
	Limit      int
	Offset     int
}
