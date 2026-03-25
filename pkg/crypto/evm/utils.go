package evm_utils

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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
