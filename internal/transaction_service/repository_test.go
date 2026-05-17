package transaction_service

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	uuid "github.com/google/uuid"
)

func TestMarkWithdrawalOperationAsProcessedReleasesBlockedBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}
	defer db.Close()

	repository := NewTransactionRepository(db, nil)
	operationIDs := []uuid.UUID{uuid.New()}

	mock.ExpectBegin()
	mock.ExpectQuery("WITH marked_operations").
		WithArgs(sqlmock.AnyArg(), "0xtx").
		WillReturnRows(sqlmock.NewRows(
			[]string{"balance_updates_count", "updated_balances_count"},
		).AddRow(1, 1))
	mock.ExpectCommit()

	err = repository.MarkWithdrawalOperationAsProcessed(operationIDs, "0xtx")
	if err != nil {
		t.Fatalf("MarkWithdrawalOperationAsProcessed() error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMarkWithdrawalOperationAsProcessedRollsBackWhenBalanceIsNotReleased(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}
	defer db.Close()

	repository := NewTransactionRepository(db, nil)
	operationIDs := []uuid.UUID{uuid.New()}

	mock.ExpectBegin()
	mock.ExpectQuery("WITH marked_operations").
		WithArgs(sqlmock.AnyArg(), "0xtx").
		WillReturnRows(sqlmock.NewRows(
			[]string{"balance_updates_count", "updated_balances_count"},
		).AddRow(1, 0))
	mock.ExpectRollback()

	err = repository.MarkWithdrawalOperationAsProcessed(operationIDs, "0xtx")
	if err == nil {
		t.Fatal("expected balance release mismatch to return an error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
