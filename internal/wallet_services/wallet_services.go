package walletservices

import (
	hex "encoding/hex"

	system "deposit-collector/internal/system"
	crypto "deposit-collector/pkg/crypto/wallet"
	logger "deposit-collector/pkg/logger"
	utils "deposit-collector/pkg/utils"
)

type Wallet struct {
	Address    string
	PrivateKey string
	PublicKey  string
	Path       string
	WIF        string
}

type WalletServices struct {
	seed []byte
}

func (s *WalletServices) GenerateWallet(
	accountIndex uint32,
	changeIndex uint32,
	index uint32,
	chain system.ChainPlatform,
) (crypto.CryptoWallet, error) {
	switch chain {
	case system.ChainPlatformEVM:
		evmWallet, err := crypto.GenerateEvmWallet(
			s.seed, accountIndex, changeIndex, index,
		)
		if err != nil {
			return nil, err
		}
		return evmWallet, nil
	case system.ChainPlatformBTC:
		btcWallet, err := crypto.GenerateBitcoinWallet(
			s.seed,
			false,
			crypto.PurposeBTCNativeSegwit,
			accountIndex,
			changeIndex,
			index,
		)
		if err != nil {
			return nil, err
		}
		return btcWallet, nil
	case system.ChainPlatformSOL:
		solWallet, err := crypto.GenerateSolanaWallet(
			s.seed, accountIndex, changeIndex, index,
		)
		if err != nil {
			return nil, err
		}
		return solWallet, nil
	default:
		return nil, utils.NewError("invalid chain platform")
	}
}

func NewWalletServices(
	seed string,
	logger *logger.Logger,
) *WalletServices {
	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		utils.FailOnError(
			logger,
			err,
			"Error decoding wallet seed",
		)
	}
	return &WalletServices{
		seed: seedBytes,
	}
}
