package service

import (
	"errors"
	"math"
	"testing"

	"simple_calculator/internal/domain"
)

func TestAdd(t *testing.T) {
	calculator := NewCalculator()
	tests := []struct {
		name     string
		operands []float64
		want     float64
		wantErr  error
	}{
		{"integers", []float64{10, 4, 2}, 16, nil},
		{"decimals", []float64{1.5, 2.25}, 3.75, nil},
		{"negative", []float64{-5, 2}, -3, nil},
		{"zero", []float64{0, 0}, 0, nil},
		{"missing operands", nil, 0, domain.ErrInsufficientOperands},
		{"one operand", []float64{1}, 0, domain.ErrInsufficientOperands},
		{"overflow", []float64{math.MaxFloat64, math.MaxFloat64}, 0, domain.ErrNonRepresentableResult},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := calculator.Add(test.operands)
			assertResult(t, got, err, test.want, test.wantErr)
		})
	}
}

func TestMultiply(t *testing.T) {
	calculator := NewCalculator()
	tests := []struct {
		name     string
		operands []float64
		want     float64
		wantErr  error
	}{
		{"normal", []float64{10, 4, 2}, 80, nil},
		{"decimals", []float64{1.5, 2}, 3, nil},
		{"negative", []float64{-5, 2}, -10, nil},
		{"zero", []float64{4, 0}, 0, nil},
		{"one operand", []float64{1}, 0, domain.ErrInsufficientOperands},
		{"overflow", []float64{math.MaxFloat64, 2}, 0, domain.ErrNonRepresentableResult},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := calculator.Multiply(test.operands)
			assertResult(t, got, err, test.want, test.wantErr)
		})
	}
}

func TestDivide(t *testing.T) {
	calculator := NewCalculator()
	tests := []struct {
		name     string
		operands []float64
		want     float64
		wantErr  error
	}{
		{"left to right", []float64{10, 4, 2}, 1.25, nil},
		{"decimals", []float64{7.5, 2.5}, 3, nil},
		{"negative", []float64{-8, 2}, -4, nil},
		{"zero dividend", []float64{0, 10, 10}, 0, nil},
		{"zero middle divisor", []float64{10, 0, 4}, 0, domain.ErrDivisionByZero},
		{"zero final divisor", []float64{10, 4, 0}, 0, domain.ErrDivisionByZero},
		{"one operand", []float64{1}, 0, domain.ErrInsufficientOperands},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := calculator.Divide(test.operands)
			assertResult(t, got, err, test.want, test.wantErr)
		})
	}
}

func TestExponentiate(t *testing.T) {
	calculator := NewCalculator()
	tests := []struct {
		name           string
		base, exponent float64
		want           float64
		wantErr        error
	}{
		{"positive", 10, 2, 100, nil},
		{"negative exponent", 2, -2, 0.25, nil},
		{"negative base", -2, 3, -8, nil},
		{"zero", 0, 2, 0, nil},
		{"NaN result", -2, 0.5, 0, domain.ErrNoRealResult},
		{"infinite result", 10, 309, 0, domain.ErrNonRepresentableResult},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := calculator.Exponentiate(test.base, test.exponent)
			assertResult(t, got, err, test.want, test.wantErr)
		})
	}
}

func TestSquareRoot(t *testing.T) {
	calculator := NewCalculator()
	tests := []struct {
		name           string
		radicand, want float64
		wantErr        error
	}{
		{"positive", 25, 5, nil},
		{"decimal", 2.25, 1.5, nil},
		{"zero", 0, 0, nil},
		{"negative", -1, 0, domain.ErrNegativeRadicand},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := calculator.SquareRoot(test.radicand)
			assertResult(t, got, err, test.want, test.wantErr)
		})
	}
}

func TestPercentage(t *testing.T) {
	calculator := NewCalculator()
	tests := []struct {
		name                    string
		value, percentage, want float64
		wantErr                 error
	}{
		{"normal", 100, 10, 10, nil},
		{"decimal", 12.5, 20, 2.5, nil},
		{"above one hundred", 50, 150, 75, nil},
		{"negative", 100, -10, -10, nil},
		{"zero", 100, 0, 0, nil},
		{"overflow", math.MaxFloat64, 200, 0, domain.ErrNonRepresentableResult},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := calculator.Percentage(test.value, test.percentage)
			assertResult(t, got, err, test.want, test.wantErr)
		})
	}
}

func TestRejectsNonFiniteInputs(t *testing.T) {
	calculator := NewCalculator()
	if _, err := calculator.Add([]float64{1, math.NaN()}); !errors.Is(err, domain.ErrNonFiniteInput) {
		t.Fatalf("Add() error = %v, want %v", err, domain.ErrNonFiniteInput)
	}
	if _, err := calculator.SquareRoot(math.Inf(1)); !errors.Is(err, domain.ErrNonFiniteInput) {
		t.Fatalf("SquareRoot() error = %v, want %v", err, domain.ErrNonFiniteInput)
	}
}

func assertResult(t *testing.T, got float64, err error, want float64, wantErr error) {
	t.Helper()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
	if wantErr == nil && math.Abs(got-want) > 1e-12 {
		t.Fatalf("result = %v, want %v", got, want)
	}
}
