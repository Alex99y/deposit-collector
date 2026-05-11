package evm_utils

import (
	errors "errors"
	big "math/big"

	common "github.com/ethereum/go-ethereum/common"
	types "github.com/ethereum/go-ethereum/core/types"
)

// UnsignedEIP1559Tx is an unsigned EIP-1559 (DynamicFeeTx) transfer payload.
// We keep the `From` address because it is needed to validate the signing key.
type UnsignedEIP1559Tx struct {
	From common.Address
	Tx   *types.DynamicFeeTx
}

// BuildNativeTransferEIP1559Tx builds an EIP-1559 native transfer transaction.
// Fee policy: maxFeePerGas = baseFeePerGas*2 + maxPriorityFeePerGas.
func BuildNativeTransferEIP1559Tx(
	chainID int,
	from common.Address,
	nonce uint64,
	to common.Address,
	valueWei *big.Int,
	gasLimit uint64,
	maxPriorityFeePerGas *big.Int,
	baseFeePerGas *big.Int,
) (*UnsignedEIP1559Tx, error) {
	if valueWei == nil || valueWei.Sign() <= 0 {
		return nil, errors.New("valueWei must be > 0")
	}
	if gasLimit == 0 {
		return nil, errors.New("gasLimit must be > 0")
	}
	if maxPriorityFeePerGas == nil || maxPriorityFeePerGas.Sign() <= 0 {
		return nil, errors.New("maxPriorityFeePerGas must be > 0")
	}
	if baseFeePerGas == nil || baseFeePerGas.Sign() <= 0 {
		return nil, errors.New("baseFeePerGas must be > 0")
	}

	feeCap := new(big.Int).Mul(baseFeePerGas, big.NewInt(2))
	feeCap.Add(feeCap, maxPriorityFeePerGas)

	chainIDBig := big.NewInt(int64(chainID))
	tx := &types.DynamicFeeTx{
		ChainID:   chainIDBig,
		Nonce:     nonce,
		To:        &to,
		Value:     valueWei,
		Gas:       gasLimit,
		GasTipCap: maxPriorityFeePerGas,
		GasFeeCap: feeCap,
	}

	return &UnsignedEIP1559Tx{
		From: from,
		Tx:   tx,
	}, nil
}
