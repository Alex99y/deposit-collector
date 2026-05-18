package evm_utils

import (
	ecdsa "crypto/ecdsa"
	hex "encoding/hex"
	errors "errors"
	fmt "fmt"
	big "math/big"
	strings "strings"

	wallet "deposit-collector/pkg/crypto/wallet"
	utils "deposit-collector/pkg/utils"

	common "github.com/ethereum/go-ethereum/common"
	types "github.com/ethereum/go-ethereum/core/types"
	crypto "github.com/ethereum/go-ethereum/crypto"
)

var transferEventSigHash = crypto.Keccak256Hash(
	[]byte("Transfer(address,address,uint256)"),
)

type ERC20Transfer struct {
	Token common.Address
	From  common.Address
	To    common.Address
	Value *big.Int
}

func ValidateAddress(address string) bool {
	return common.IsHexAddress(address)
}

func ValidateConfirmedEVMDepositReceipt(
	receipt *types.Receipt,
	latestBlockNumber uint64,
	minConfirmations int,
) error {
	if receipt == nil || receipt.BlockNumber == nil {
		return utils.NewCustomError("transaction not confirmed", false)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return utils.NewCustomError("transaction failed on-chain", false)
	}

	receiptBlock := receipt.BlockNumber.Uint64()
	if latestBlockNumber < receiptBlock {
		return utils.NewCustomError("transaction not confirmed", false)
	}

	confirmations := uint64(0)
	if minConfirmations > 0 {
		confirmations = uint64(minConfirmations)
	}
	if latestBlockNumber-receiptBlock < confirmations {
		return utils.NewCustomError("transaction not confirmed", false)
	}

	return nil
}

func FindERC20Transfers(receipt *types.Receipt) []ERC20Transfer {
	var results []ERC20Transfer

	for _, lg := range receipt.Logs {
		// topics[0] = event signature
		// topics[1] = from
		// topics[2] = to
		if len(lg.Topics) != 3 {
			continue
		}

		if lg.Topics[0] != transferEventSigHash {
			continue
		}

		fromAddress := common.BytesToAddress(lg.Topics[1].Bytes()[12:])
		toAddress := common.BytesToAddress(lg.Topics[2].Bytes()[12:])

		value := new(big.Int).SetBytes(lg.Data)
		if value.Sign() <= 0 {
			continue
		}

		results = append(results, ERC20Transfer{
			Token: lg.Address,
			From:  fromAddress,
			To:    toAddress,
			Value: value,
		})
	}

	return results
}

// EvmWalletFromWithdrawCollectorKey builds an EvmWallet from a 32-byte hex
// private key (with or without 0x prefix). Compatible with
// walletservices.SignEVMTransactions.
func EvmWalletFromWithdrawCollectorKey(
	privateKeyHex string,
) (*wallet.EvmWallet, error) {
	pk := strings.TrimSpace(privateKeyHex)
	pk = strings.TrimPrefix(pk, "0x")
	if pk == "" {
		return nil, errors.New("evm private key is empty")
	}

	key, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, fmt.Errorf("evm private key: %w", err)
	}

	pub, ok := key.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid evm public key")
	}
	address := crypto.PubkeyToAddress(*pub)

	return &wallet.EvmWallet{
		Address:    address.Hex(),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(key)),
	}, nil
}
