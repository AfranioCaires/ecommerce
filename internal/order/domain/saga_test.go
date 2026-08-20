package domain

import (
	"errors"
	"testing"
	"time"
)

func TestSagaTransitions(t *testing.T) {
	saga, errorValue := NewSaga("saga-1", "order-1", "correlation-1", time.Now())
	if errorValue != nil || saga.Start(time.Now()) != nil || saga.Complete(time.Now()) != nil || saga.Status != SagaStatusCompleted {
		t.Fatalf("unexpected saga: %#v, %v", saga, errorValue)
	}
	if errorValue := saga.Compensate(time.Now()); !errors.Is(errorValue, ErrInvalidSaga) {
		t.Fatalf("expected final state rejection, received %v", errorValue)
	}
}
