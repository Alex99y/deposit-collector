package queue

import (
	"errors"
	"testing"
)

func TestOperationConsumerArgsOperationDataReturnsUnmarshalError(t *testing.T) {
	args := OperationConsumerArgs{
		OperationEvent: OperationEvent{
			OperationType: OperationTypeDeposit,
			OperationData: []byte("{"),
		},
	}

	operation, err := args.OperationData()
	if err == nil {
		t.Fatal("expected malformed operation data to return an error")
	}
	if operation != nil {
		t.Fatalf("expected no operation, got: %#v", operation)
	}
}

func TestOperationConsumerArgsOperationDataReturnsUnknownTypeError(t *testing.T) {
	args := OperationConsumerArgs{
		OperationEvent: OperationEvent{
			OperationType: OperationType("unknown"),
			OperationData: []byte("{}"),
		},
	}

	operation, err := args.OperationData()
	if err == nil {
		t.Fatal("expected unknown operation type to return an error")
	}
	if !errors.Is(err, ErrUnknownOperationType) {
		t.Fatalf("expected ErrUnknownOperationType, got: %v", err)
	}
	if operation != nil {
		t.Fatalf("expected no operation, got: %#v", operation)
	}
}
