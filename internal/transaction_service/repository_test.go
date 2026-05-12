package transaction_service

import (
	"strings"
	"testing"
)

func TestUpdateWithdrawUserBalanceQuerySeparatesSetAssignments(t *testing.T) {
	want := strings.Join([]string{
		"available_balance = available_balance - $1,",
		"blocked_balance_for_withdrawal = blocked_balance_for_withdrawal + $1,",
		"updated_at = CURRENT_TIMESTAMP",
	}, "\n")

	if !strings.Contains(updateWithdrawUserBalanceQuery, want) {
		t.Fatalf("withdraw balance update query is missing comma-separated SET assignments:\n%s", updateWithdrawUserBalanceQuery)
	}
}
