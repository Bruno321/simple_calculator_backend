package http

import (
	"encoding/json"
	"errors"
	"io"
	stdhttp "net/http"

	"simple_calculator/internal/domain"
	"simple_calculator/internal/model"
)

// Calculator defines the service behavior required by the HTTP transport.
type Calculator interface {
	Add([]float64) (float64, error)
	Multiply([]float64) (float64, error)
	Divide([]float64) (float64, error)
	Exponentiate(float64, float64) (float64, error)
	SquareRoot(float64) (float64, error)
	Percentage(float64, float64) (float64, error)
}

type Handler struct {
	calculator Calculator
}

func NewHandler(calculator Calculator) *Handler {
	return &Handler{calculator: calculator}
}

func (h *Handler) addition(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handleJSON(w, r, validateOperandsRequest, func(request model.OperandRequest) (float64, error) {
		return h.calculator.Add(operandValues(request))
	})
}

func (h *Handler) multiplication(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handleJSON(w, r, validateOperandsRequest, func(request model.OperandRequest) (float64, error) {
		return h.calculator.Multiply(operandValues(request))
	})
}

func (h *Handler) division(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handleJSON(w, r, validateOperandsRequest, func(request model.OperandRequest) (float64, error) {
		return h.calculator.Divide(operandValues(request))
	})
}

func (h *Handler) exponentiation(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handleJSON(w, r, validateExponentiationRequest, func(request model.ExponentiationRequest) (float64, error) {
		return h.calculator.Exponentiate(*request.Base, *request.Exponent)
	})
}

func (h *Handler) squareRoot(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handleJSON(w, r, validateSquareRootRequest, func(request model.SquareRootRequest) (float64, error) {
		return h.calculator.SquareRoot(*request.Radicand)
	})
}

func (h *Handler) percentage(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	handleJSON(w, r, validatePercentageRequest, func(request model.PercentageRequest) (float64, error) {
		return h.calculator.Percentage(*request.Value, *request.Percentage)
	})
}

func handleJSON[T any](
	w stdhttp.ResponseWriter,
	r *stdhttp.Request,
	validate func(T) error,
	calculate func(T) (float64, error),
) {
	var request T
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, stdhttp.StatusBadRequest, err.Error())
		return
	}
	if err := validate(request); err != nil {
		writeError(w, stdhttp.StatusBadRequest, err.Error())
		return
	}

	result, err := calculate(request)
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, stdhttp.StatusOK, model.ResultResponse{Result: result})
}

func decodeJSON(r *stdhttp.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return errors.New("invalid JSON")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("invalid JSON: request body must contain one JSON object")
	}
	return nil
}

func validateOperandsRequest(request model.OperandRequest) error {
	if request.Operands == nil {
		return errors.New("operands is required")
	}
	for _, operand := range *request.Operands {
		if operand == nil {
			return errors.New("operands must contain only numbers")
		}
	}
	return nil
}

func operandValues(request model.OperandRequest) []float64 {
	values := make([]float64, len(*request.Operands))
	for index, operand := range *request.Operands {
		values[index] = *operand
	}
	return values
}

func validateExponentiationRequest(request model.ExponentiationRequest) error {
	if request.Base == nil {
		return errors.New("base is required")
	}
	if request.Exponent == nil {
		return errors.New("exponent is required")
	}
	return nil
}

func validateSquareRootRequest(request model.SquareRootRequest) error {
	if request.Radicand == nil {
		return errors.New("radicand is required")
	}
	return nil
}

func validatePercentageRequest(request model.PercentageRequest) error {
	if request.Value == nil {
		return errors.New("value is required")
	}
	if request.Percentage == nil {
		return errors.New("percentage is required")
	}
	return nil
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, domain.ErrInsufficientOperands),
		errors.Is(err, domain.ErrDivisionByZero),
		errors.Is(err, domain.ErrNegativeRadicand),
		errors.Is(err, domain.ErrNoRealResult),
		errors.Is(err, domain.ErrNonFiniteInput),
		errors.Is(err, domain.ErrNonRepresentableResult):
		return stdhttp.StatusBadRequest
	default:
		return stdhttp.StatusInternalServerError
	}
}
