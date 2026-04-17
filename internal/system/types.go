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
	ChainName     string        `json:"chainName"`
	ChainPlatform ChainPlatform `json:"chainPlatform"`
	EVMChainID    *int          `json:"evmChainId,omitempty"`
}

type SupportedChain struct {
	ChainDbID uuid.UUID `json:"chainDbId"`
	NewSupportedChainRequest
}

/*
*
ChainName is the name of the chain.
Example: Ethereum
*/

type BaseTokenAddress struct {
	UnitName   string `json:"unitName"`
	UnitSymbol string `json:"unitSymbol"`
	Address    string `json:"address"`
	Decimals   int    `json:"decimals"`
}

type NewTokenAddressRequest struct {
	BaseTokenAddress
	ChainName string `json:"chainName"`
}

type TokenAddress struct {
	TokenAddressDbID uuid.UUID `json:"tokenAddressDbId"`
	BaseTokenAddress
	Chain SupportedChain `json:"supportedChain"`
}

type GetTokenAddressesRequest struct {
	Chain      *string
	Address    *string
	UnitSymbol *string
	Limit      int
	Offset     int
}
