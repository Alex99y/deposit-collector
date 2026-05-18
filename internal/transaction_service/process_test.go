package transaction_service

import (
	big "math/big"
	testing "testing"

	evm_utils "deposit-collector/pkg/crypto/evm"
	utils "deposit-collector/pkg/utils"

	types "github.com/ethereum/go-ethereum/core/types"
)

func TestValidateConfirmedEVMDepositReceiptRejectsFailedReceipt(t *testing.T) {
	err := evm_utils.ValidateConfirmedEVMDepositReceipt(
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
	customError, ok := utils.IsCustomError(err)
	if !ok {
		t.Fatalf("expected CustomError, got %T", err)
	}
	if customError.IsRetryable() {
		t.Fatal("expected failed receipt to be non-retryable")
	}
}

func TestValidateConfirmedEVMDepositReceiptRequiresConfirmationDepth(t *testing.T) {
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(10),
	}

	err := evm_utils.ValidateConfirmedEVMDepositReceipt(receipt, 14, 5)
	if err == nil {
		t.Fatal("expected receipt below confirmation depth to return an error")
	}

	err = evm_utils.ValidateConfirmedEVMDepositReceipt(receipt, 15, 5)
	if err != nil {
		t.Fatalf("validateConfirmedEVMDepositReceipt() unexpected error: %v", err)
	}
}

func TestValidateConfirmedEVMDepositReceiptRejectsMissingReceipt(t *testing.T) {
	err := evm_utils.ValidateConfirmedEVMDepositReceipt(nil, 20, 5)
	if err == nil {
		t.Fatal("expected missing receipt to return an error")
	}
}
