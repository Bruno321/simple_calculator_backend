package domain

import "errors"

var (
	ErrInsufficientOperands   = errors.New("at least two operands are required")
	ErrDivisionByZero         = errors.New("division by zero")
	ErrNegativeRadicand       = errors.New("square root of a negative number has no real result")
	ErrNoRealResult           = errors.New("calculation has no real-number result")
	ErrNonFiniteInput         = errors.New("values must be finite real numbers")
	ErrNonRepresentableResult = errors.New("result cannot be represented as a finite number")
)
