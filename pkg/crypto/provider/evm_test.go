package provider

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestConfirmedSuccessfulReceiptRejectsFailedReceipt(t *testing.T) {
	confirmed, err := confirmedSuccessfulReceipt(
		&types.Receipt{
			Status:      types.ReceiptStatusFailed,
			BlockNumber: big.NewInt(10),
		},
		20,
		5,
	)

	if err == nil {
		t.Fatal("expected failed receipt to return an error")
	}
	if confirmed {
		t.Fatal("expected failed receipt to be unconfirmed")
	}
}

func TestConfirmedSuccessfulReceiptRequiresConfirmationDepth(t *testing.T) {
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(10),
	}

	confirmed, err := confirmedSuccessfulReceipt(receipt, 14, 5)
	if err != nil {
		t.Fatalf("confirmedSuccessfulReceipt() unexpected error: %v", err)
	}
	if confirmed {
		t.Fatal("expected receipt below confirmation depth to be unconfirmed")
	}

	confirmed, err = confirmedSuccessfulReceipt(receipt, 15, 5)
	if err != nil {
		t.Fatalf("confirmedSuccessfulReceipt() unexpected error: %v", err)
	}
	if !confirmed {
		t.Fatal("expected successful receipt at confirmation depth to be confirmed")
	}
}

func TestTransactionRecipientHexHandlesContractCreation(t *testing.T) {
	tx := types.NewContractCreation(
		1,
		big.NewInt(1),
		21_000,
		big.NewInt(1),
		[]byte{0x60, 0x00},
	)

	if got := transactionRecipientHex(tx); got != "" {
		t.Fatalf("recipient = %q, want empty for contract creation", got)
	}
}

func TestTransactionRecipientHexReturnsRecipient(t *testing.T) {
	recipient := common.HexToAddress("0x000000000000000000000000000000000000dEaD")
	tx := types.NewTransaction(
		1,
		recipient,
		big.NewInt(1),
		21_000,
		big.NewInt(1),
		nil,
	)

	if got := transactionRecipientHex(tx); got != recipient.Hex() {
		t.Fatalf("recipient = %q, want %q", got, recipient.Hex())
	}
}
