package transaction_service

import (
	math "math/big"
	"testing"

	utils "deposit-collector/pkg/utils"

	types "github.com/ethereum/go-ethereum/core/types"
)

func TestValidateEVMDepositReceiptRejectsFailedReceiptAsNonRetryable(t *testing.T) {
	err := validateEVMDepositReceipt(
		&types.Receipt{
			Status:      types.ReceiptStatusFailed,
			BlockNumber: math.NewInt(10),
		},
		20,
		5,
	)

	customError, ok := utils.IsCustomError(err)
	if !ok {
		t.Fatalf("expected custom error, got: %v", err)
	}
	if customError.IsRetryable() {
		t.Fatal("expected failed on-chain transaction to be non-retryable")
	}
}

func TestValidateEVMDepositReceiptTreatsMissingDepthAsRetryable(t *testing.T) {
	err := validateEVMDepositReceipt(
		&types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			BlockNumber: math.NewInt(10),
		},
		14,
		5,
	)

	customError, ok := utils.IsCustomError(err)
	if !ok {
		t.Fatalf("expected custom error, got: %v", err)
	}
	if !customError.IsRetryable() {
		t.Fatal("expected unconfirmed transaction to be retryable")
	}
}

func TestValidateEVMDepositReceiptAcceptsSuccessfulReceiptAtDepth(t *testing.T) {
	err := validateEVMDepositReceipt(
		&types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			BlockNumber: math.NewInt(10),
		},
		15,
		5,
	)

	if err != nil {
		t.Fatalf("expected receipt to be accepted, got: %v", err)
	}
}

func TestERC20TransferAmountToInt64RejectsOverflow(t *testing.T) {
	overflowingAmount := new(math.Int).Lsh(math.NewInt(1), 63)

	_, err := erc20TransferAmountToInt64(overflowingAmount)

	customError, ok := utils.IsCustomError(err)
	if !ok {
		t.Fatalf("expected custom error, got: %v", err)
	}
	if customError.IsRetryable() {
		t.Fatal("expected unsupported ERC-20 amount to be non-retryable")
	}
}

func TestERC20TransferAmountToInt64AcceptsSupportedAmount(t *testing.T) {
	amount, err := erc20TransferAmountToInt64(math.NewInt(123))

	if err != nil {
		t.Fatalf("expected amount to be accepted, got: %v", err)
	}
	if amount != 123 {
		t.Fatalf("expected amount 123, got: %d", amount)
	}
}
