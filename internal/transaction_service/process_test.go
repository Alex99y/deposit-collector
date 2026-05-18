package transaction_service

import (
	"math"
	"math/big"
	"strings"
	"testing"

	queue "deposit-collector/internal/queue"
	system "deposit-collector/internal/system"
	provider "deposit-collector/pkg/crypto/provider"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestProcessEVMDepositTxInfoRejectsFailedReceipt(t *testing.T) {
	targetAddress := common.HexToAddress("0x1111111111111111111111111111111111111111").Hex()
	operation := &queue.DepositOperationEvent{
		TargetChainName: "ethereum",
		TargetAddress:   targetAddress,
		DepositTxHash:   "0xfailed",
	}

	processedOperation, err := processEVMDepositTxInfo(
		&provider.EvmTxInfo{
			To:     targetAddress,
			Amount: "1000",
			TxReceipt: &types.Receipt{
				Status:      types.ReceiptStatusFailed,
				BlockNumber: big.NewInt(10),
			},
		},
		20,
		5,
		operation,
		func(string, string) *system.TokenAddress { return nil },
	)

	if err == nil {
		t.Fatal("expected failed receipt to be rejected")
	}
	if processedOperation != nil {
		t.Fatalf("expected no processed operation, got %+v", processedOperation)
	}
	if !strings.Contains(err.Error(), "failed on-chain") {
		t.Fatalf("expected failed receipt error, got %q", err.Error())
	}
}

func TestProcessEVMDepositTxInfoRejectsOversizedERC20Amount(t *testing.T) {
	fromAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	toAddress := common.HexToAddress("0x2222222222222222222222222222222222222222")
	tokenAddress := common.HexToAddress("0x3333333333333333333333333333333333333333")
	transferEventSigHash := crypto.Keccak256Hash(
		[]byte("Transfer(address,address,uint256)"),
	)
	overflowAmount := new(big.Int).Add(
		big.NewInt(math.MaxInt64),
		big.NewInt(1),
	)

	operation := &queue.DepositOperationEvent{
		TargetChainName: "ethereum",
		TargetAddress:   toAddress.Hex(),
		DepositTxHash:   "0xoverflow",
	}

	processedOperation, err := processEVMDepositTxInfo(
		&provider.EvmTxInfo{
			Input: []byte{0x01},
			TxReceipt: &types.Receipt{
				Status:      types.ReceiptStatusSuccessful,
				BlockNumber: big.NewInt(10),
				Logs: []*types.Log{
					{
						Address: tokenAddress,
						Topics: []common.Hash{
							transferEventSigHash,
							common.BytesToHash(fromAddress.Bytes()),
							common.BytesToHash(toAddress.Bytes()),
						},
						Data: overflowAmount.Bytes(),
					},
				},
			},
		},
		20,
		5,
		operation,
		func(string, string) *system.TokenAddress {
			return &system.TokenAddress{}
		},
	)

	if err == nil {
		t.Fatal("expected oversized ERC20 amount to be rejected")
	}
	if processedOperation != nil {
		t.Fatalf("expected no processed operation, got %+v", processedOperation)
	}
	if !strings.Contains(err.Error(), "exceeds int64") {
		t.Fatalf("expected overflow error, got %q", err.Error())
	}
}
