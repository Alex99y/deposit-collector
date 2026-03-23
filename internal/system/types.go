package system

import (
	errors "errors"

	uuid "github.com/google/uuid"
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
	ChainName     string
	ChainPlatform ChainPlatform
	EVMChainID    int
}

type SupportedChain struct {
	ChainDbID uuid.UUID
	NewSupportedChainRequest
}

/*
*
ChainName is the name of the chain.
Example: Ethereum
*/

type BaseTokenAddress struct {
	UnitName   string
	UnitSymbol string
	Address    string
	Decimals   int
}

type NewTokenAddressRequest struct {
	BaseTokenAddress
	ChainName string
}

type TokenAddress struct {
	TokenAddressDbID uuid.UUID
	BaseTokenAddress
	Chain SupportedChain
}

type GetTokenAddressesRequest struct {
	Chain      *string
	Address    *string
	UnitSymbol *string
	Limit      int
	Offset     int
}
