package evm_utils

import (
	bytes "bytes"
	big "math/big"
	testing "testing"

	common "github.com/ethereum/go-ethereum/common"
)

func TestCalculateNativeSweepValueReservesMaxGasCost(t *testing.T) {
	balanceWei := big.NewInt(1_000_000_000_000_000_000)
	gasFeeCap := big.NewInt(3_000_000_000)

	sweepValue, err := CalculateNativeSweepValue(
		balanceWei,
		NativeTransferIntrinsicGas,
		gasFeeCap,
	)
	if err != nil {
		t.Fatalf("CalculateNativeSweepValue() unexpected error: %v", err)
	}

	want := big.NewInt(999_937_000_000_000_000)
	if sweepValue.Cmp(want) != 0 {
		t.Fatalf("sweep value = %s, want %s", sweepValue, want)
	}
	if balanceWei.String() != "1000000000000000000" {
		t.Fatalf("balanceWei was mutated: %s", balanceWei)
	}
	if gasFeeCap.String() != "3000000000" {
		t.Fatalf("gasFeeCap was mutated: %s", gasFeeCap)
	}
}

func TestCalculateNativeSweepValueRejectsBalanceConsumedByGas(t *testing.T) {
	_, err := CalculateNativeSweepValue(
		big.NewInt(63_000_000_000_000),
		NativeTransferIntrinsicGas,
		big.NewInt(3_000_000_000),
	)
	if err == nil {
		t.Fatal("expected error when gas consumes the full balance")
	}
}

func TestCalculateEIP1559GasFeeCap(t *testing.T) {
	baseFeePerGas := big.NewInt(10)
	tipCap := big.NewInt(2)

	feeCap, err := CalculateEIP1559GasFeeCap(tipCap, baseFeePerGas)
	if err != nil {
		t.Fatalf("CalculateEIP1559GasFeeCap() unexpected error: %v", err)
	}
	if feeCap.Int64() != 22 {
		t.Fatalf("fee cap = %d, want 22", feeCap.Int64())
	}
	if baseFeePerGas.Int64() != 10 {
		t.Fatalf("baseFeePerGas was mutated: %d", baseFeePerGas.Int64())
	}
}

func TestEncodeERC20TransferLengthAndSelector(t *testing.T) {
	recipient := common.HexToAddress("0x000000000000000000000000000000000000dEaD")
	amount := big.NewInt(1_000_000)

	data, err := EncodeERC20Transfer(recipient, amount)
	if err != nil {
		t.Fatalf("EncodeERC20Transfer() unexpected error: %v", err)
	}
	if len(data) != 4+32+32 {
		t.Fatalf("calldata len = %d, want 68", len(data))
	}
	// transfer(address,uint256) selector
	want := []byte{0xa9, 0x05, 0x9c, 0xbb}
	if !bytes.Equal(data[:4], want) {
		t.Fatalf("selector = %x, want a9059cbb", data[:4])
	}
}
