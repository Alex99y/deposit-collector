package transaction_service

import (
	"math"
	"math/big"
	"strings"
	"testing"
)

func TestERC20TransferAmountToInt64(t *testing.T) {
	t.Run("accepts maximum supported amount", func(t *testing.T) {
		amount, err := erc20TransferAmountToInt64(big.NewInt(math.MaxInt64))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if amount != math.MaxInt64 {
			t.Fatalf("expected %d, got %d", int64(math.MaxInt64), amount)
		}
	})

	t.Run("rejects amount beyond storage range", func(t *testing.T) {
		oversizedAmount := new(big.Int).Add(big.NewInt(math.MaxInt64), big.NewInt(1))

		amount, err := erc20TransferAmountToInt64(oversizedAmount)
		if err == nil {
			t.Fatalf("expected error, got amount %d", amount)
		}
		if !strings.Contains(err.Error(), "exceeds supported range") {
			t.Fatalf("expected range error, got %v", err)
		}
	})
}
