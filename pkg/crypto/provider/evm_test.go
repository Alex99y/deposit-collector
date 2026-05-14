package provider

import (
	"math/big"
	"testing"

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
