package system

import "github.com/google/uuid"

type ChainPlatform string

const (
	ChainPlatformEVM ChainPlatform = "EVM"
	ChainPlatformBTC ChainPlatform = "BTC"
	ChainPlatformSOL ChainPlatform = "SOL"
)

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
Chain is the network of the chain.
Example: Ethereum
*/
type NewTokenAddressRequest struct {
	UnitName   string
	UnitSymbol string
	Address    string
	Chain      string
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
