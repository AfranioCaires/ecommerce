package domain

import (
	"errors"
	"time"
)

type SagaStatus string

const (
	SagaStatusPending     SagaStatus = "PENDING"
	SagaStatusProcessing  SagaStatus = "PROCESSING"
	SagaStatusCompleted   SagaStatus = "COMPLETED"
	SagaStatusCompensated SagaStatus = "COMPENSATED"
)

var ErrInvalidSaga = errors.New("the order saga is invalid.")

type Saga struct {
	ID, OrderID, CorrelationID string
	Status                     SagaStatus
	CreatedAt, UpdatedAt       time.Time
}

func NewSaga(id, orderID, correlationID string, createdAt time.Time) (*Saga, error) {
	if id == "" || orderID == "" || correlationID == "" || createdAt.IsZero() {
		return nil, ErrInvalidSaga
	}
	normalized := createdAt.UTC()
	return &Saga{ID: id, OrderID: orderID, CorrelationID: correlationID, Status: SagaStatusPending, CreatedAt: normalized, UpdatedAt: normalized}, nil
}

func RestoreSaga(id, orderID, correlationID string, status SagaStatus, createdAt, updatedAt time.Time) (*Saga, error) {
	saga, errorValue := NewSaga(id, orderID, correlationID, createdAt)
	if errorValue != nil || !status.valid() {
		return nil, ErrInvalidSaga
	}
	saga.Status, saga.UpdatedAt = status, updatedAt.UTC()
	return saga, nil
}

func (saga *Saga) Start(at time.Time) error {
	return saga.transition(SagaStatusPending, SagaStatusProcessing, at)
}
func (saga *Saga) Complete(at time.Time) error {
	return saga.transition(SagaStatusProcessing, SagaStatusCompleted, at)
}
func (saga *Saga) Compensate(at time.Time) error {
	return saga.transition(SagaStatusProcessing, SagaStatusCompensated, at)
}
func (saga *Saga) transition(from, to SagaStatus, at time.Time) error {
	if saga.Status != from {
		return ErrInvalidSaga
	}
	saga.Status, saga.UpdatedAt = to, at.UTC()
	return nil
}
func (status SagaStatus) valid() bool {
	return status == SagaStatusPending || status == SagaStatusProcessing || status == SagaStatusCompleted || status == SagaStatusCompensated
}
