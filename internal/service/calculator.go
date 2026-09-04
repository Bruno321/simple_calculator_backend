package service

import (
	"math"

	"simple_calculator/internal/domain"
)

// Calculator contains the calculator's mathematical rules and has no HTTP dependencies.
type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Add(operands []float64) (float64, error) {
	if err := validateOperands(operands); err != nil {
		return 0, err
	}

	result := operands[0]
	for _, operand := range operands[1:] {
		result += operand
		if !isFinite(result) {
			return 0, domain.ErrNonRepresentableResult
		}
	}
	return result, nil
}

func (c *Calculator) Subtract(operands []float64) (float64, error) {
	if err := validateOperands(operands); err != nil {
		return 0, err
	}

	result := operands[0]
	for _, operand := range operands[1:] {
		result -= operand
		if !isFinite(result) {
			return 0, domain.ErrNonRepresentableResult
		}
	}
	return result, nil
}

func (c *Calculator) Multiply(operands []float64) (float64, error) {
	if err := validateOperands(operands); err != nil {
		return 0, err
	}

	result := operands[0]
	for _, operand := range operands[1:] {
		result *= operand
		if !isFinite(result) {
			return 0, domain.ErrNonRepresentableResult
		}
	}
	return result, nil
}

func (c *Calculator) Divide(operands []float64) (float64, error) {
	if err := validateOperands(operands); err != nil {
		return 0, err
	}

	result := operands[0]
	for _, divisor := range operands[1:] {
		if divisor == 0 {
			return 0, domain.ErrDivisionByZero
		}
		result /= divisor
		if !isFinite(result) {
			return 0, domain.ErrNonRepresentableResult
		}
	}
	return result, nil
}

func (c *Calculator) Exponentiate(base, exponent float64) (float64, error) {
	if !isFinite(base) || !isFinite(exponent) {
		return 0, domain.ErrNonFiniteInput
	}

	result := math.Pow(base, exponent)
	if math.IsNaN(result) {
		return 0, domain.ErrNoRealResult
	}
	if math.IsInf(result, 0) {
		return 0, domain.ErrNonRepresentableResult
	}
	return result, nil
}

func (c *Calculator) SquareRoot(radicand float64) (float64, error) {
	if !isFinite(radicand) {
		return 0, domain.ErrNonFiniteInput
	}
	if radicand < 0 {
		return 0, domain.ErrNegativeRadicand
	}
	return math.Sqrt(radicand), nil
}

func (c *Calculator) Percentage(value, percentage float64) (float64, error) {
	if !isFinite(value) || !isFinite(percentage) {
		return 0, domain.ErrNonFiniteInput
	}

	result := value * percentage
	if isFinite(result) {
		result /= 100
	} else {
		// Scaling first can avoid an overflowing intermediate when the final
		// percentage result is still representable.
		result = value * (percentage / 100)
	}
	if !isFinite(result) {
		return 0, domain.ErrNonRepresentableResult
	}
	return result, nil
}

func validateOperands(operands []float64) error {
	if len(operands) < 2 {
		return domain.ErrInsufficientOperands
	}
	for _, operand := range operands {
		if !isFinite(operand) {
			return domain.ErrNonFiniteInput
		}
	}
	return nil
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
