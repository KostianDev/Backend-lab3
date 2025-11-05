package storage

import "errors"

var (
	// ErrNotFound is returned when an entity cannot be located.
	ErrNotFound = errors.New("not found")
	// ErrConflict signals that a conflicting entity already exists.
	ErrConflict = errors.New("conflict")
	// ErrPreconditionFailed indicates a business rule guard prevented an operation.
	ErrPreconditionFailed = errors.New("precondition failed")
	// ErrInsufficientFunds indicates an expense would drive balance below zero while forbidden.
	ErrInsufficientFunds = errors.New("insufficient funds")
)
