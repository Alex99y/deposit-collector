package crypto

import (
	system "deposit-collector/internal/system"
	btc_utils "deposit-collector/pkg/crypto/btc"
	evm_utils "deposit-collector/pkg/crypto/evm"
)

type CryptoUtils struct {
	bitcoinNetwork btc_utils.NETWORK
}

func (c *CryptoUtils) GetBitcoinNetwork() btc_utils.NETWORK {
	return c.bitcoinNetwork
}

func (c *CryptoUtils) ValidateAddress(
	address string,
	chain system.ChainPlatform,
) bool {
	switch chain {
	case system.ChainPlatformEVM:
		return evm_utils.ValidateAddress(address)
	case system.ChainPlatformBTC:
		return btc_utils.ValidateAddress(address, c.bitcoinNetwork)
	default:
		return false
	}
}

func NewCryptoUtils(bitcoinNetwork btc_utils.NETWORK) *CryptoUtils {
	return &CryptoUtils{
		bitcoinNetwork: bitcoinNetwork,
	}
}
