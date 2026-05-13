package evm_utils

import (
	errors "errors"
	big "math/big"

	common "github.com/ethereum/go-ethereum/common"
	types "github.com/ethereum/go-ethereum/core/types"
	crypto "github.com/ethereum/go-ethereum/crypto"
)

const NativeTransferIntrinsicGas uint64 = 21000

// UnsignedEIP1559Tx is an unsigned EIP-1559 (DynamicFeeTx) transfer payload.
// We keep the `From` address because it is needed to validate the signing key.
type UnsignedEIP1559Tx struct {
	From common.Address
	Tx   *types.DynamicFeeTx
}

// CalculateEIP1559GasFeeCap returns the max fee per gas used by the collector's
// fee policy: maxFeePerGas = baseFeePerGas*2 + maxPriorityFeePerGas.
func CalculateEIP1559GasFeeCap(
	maxPriorityFeePerGas *big.Int,
	baseFeePerGas *big.Int,
) (*big.Int, error) {
	if maxPriorityFeePerGas == nil || maxPriorityFeePerGas.Sign() <= 0 {
		return nil, errors.New("maxPriorityFeePerGas must be > 0")
	}
	if baseFeePerGas == nil || baseFeePerGas.Sign() <= 0 {
		return nil, errors.New("baseFeePerGas must be > 0")
	}

	feeCap := new(big.Int).Mul(new(big.Int).Set(baseFeePerGas), big.NewInt(2))
	feeCap.Add(feeCap, maxPriorityFeePerGas)
	return feeCap, nil
}

// CalculateNativeSweepValue subtracts the maximum gas reservation from a native
// token balance so a sweep transaction can be funded by the source account.
func CalculateNativeSweepValue(
	balanceWei *big.Int,
	gasLimit uint64,
	gasFeeCap *big.Int,
) (*big.Int, error) {
	if balanceWei == nil || balanceWei.Sign() <= 0 {
		return nil, errors.New("balanceWei must be > 0")
	}
	if gasLimit == 0 {
		return nil, errors.New("gasLimit must be > 0")
	}
	if gasFeeCap == nil || gasFeeCap.Sign() <= 0 {
		return nil, errors.New("gasFeeCap must be > 0")
	}

	maxGasCost := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), gasFeeCap)
	sweepValue := new(big.Int).Sub(new(big.Int).Set(balanceWei), maxGasCost)
	if sweepValue.Sign() <= 0 {
		return nil, errors.New("balance is not enough to cover gas")
	}
	return sweepValue, nil
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
	feeCap, err := CalculateEIP1559GasFeeCap(
		maxPriorityFeePerGas, baseFeePerGas,
	)
	if err != nil {
		return nil, err
	}

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

// EncodeERC20Transfer builds calldata for ERC-20 transfer(address,uint256).
func EncodeERC20Transfer(recipient common.Address, amount *big.Int) ([]byte, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, errors.New("transfer amount must be > 0")
	}
	methodID := crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]
	paddedAddr := common.LeftPadBytes(recipient.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	return append(append(methodID, paddedAddr...), paddedAmount...), nil
}

// BuildContractCallEIP1559Tx builds an EIP-1559 transaction calling a contract
// (native value optional, usually zero for ERC-20 transfers).
func BuildContractCallEIP1559Tx(
	chainID int,
	from common.Address,
	nonce uint64,
	contract common.Address,
	valueWei *big.Int,
	data []byte,
	gasLimit uint64,
	maxPriorityFeePerGas *big.Int,
	baseFeePerGas *big.Int,
) (*UnsignedEIP1559Tx, error) {
	if gasLimit == 0 {
		return nil, errors.New("gasLimit must be > 0")
	}
	value := valueWei
	if value == nil {
		value = big.NewInt(0)
	}
	feeCap, err := CalculateEIP1559GasFeeCap(
		maxPriorityFeePerGas, baseFeePerGas,
	)
	if err != nil {
		return nil, err
	}

	chainIDBig := big.NewInt(int64(chainID))
	tx := &types.DynamicFeeTx{
		ChainID:   chainIDBig,
		Nonce:     nonce,
		To:        &contract,
		Value:     new(big.Int).Set(value),
		Gas:       gasLimit,
		GasTipCap: maxPriorityFeePerGas,
		GasFeeCap: feeCap,
		Data:      data,
	}

	return &UnsignedEIP1559Tx{
		From: from,
		Tx:   tx,
	}, nil
}
