package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simple_calculator/internal/domain"
)

func TestCalculatorEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		serviceErr error
		wantStatus int
		wantBody   string
		wantCall   string
	}{
		{"addition", "/addition", `{"operands":[10,4,2]}`, nil, 200, `{"result":42}` + "\n", "addition"},
		{"multiplication", "/multiplication", `{"operands":[1.5,4]}`, nil, 200, `{"result":42}` + "\n", "multiplication"},
		{"division", "/division", `{"operands":[10,4,2]}`, nil, 200, `{"result":42}` + "\n", "division"},
		{"exponentiation", "/exponentiation", `{"base":10,"exponent":2}`, nil, 200, `{"result":42}` + "\n", "exponentiation"},
		{"square root", "/square-root", `{"radicand":25}`, nil, 200, `{"result":42}` + "\n", "square-root"},
		{"percentage", "/percentage", `{"value":100,"percentage":10}`, nil, 200, `{"result":42}` + "\n", "percentage"},
		{"zero base", "/exponentiation", `{"base":0,"exponent":2}`, nil, 200, `{"result":42}` + "\n", "exponentiation"},
		{"zero radicand", "/square-root", `{"radicand":0}`, nil, 200, `{"result":42}` + "\n", "square-root"},
		{"zero percentage value", "/percentage", `{"value":0,"percentage":10}`, nil, 200, `{"result":42}` + "\n", "percentage"},
		{"zero percentage", "/percentage", `{"value":100,"percentage":0}`, nil, 200, `{"result":42}` + "\n", "percentage"},
		{"malformed JSON", "/addition", `{"operands":[1,2]`, nil, 400, `{"error":"invalid JSON"}` + "\n", ""},
		{"missing operands", "/addition", `{}`, nil, 400, `{"error":"operands is required"}` + "\n", ""},
		{"missing base", "/exponentiation", `{"exponent":2}`, nil, 400, `{"error":"base is required"}` + "\n", ""},
		{"missing exponent", "/exponentiation", `{"base":10}`, nil, 400, `{"error":"exponent is required"}` + "\n", ""},
		{"missing radicand", "/square-root", `{}`, nil, 400, `{"error":"radicand is required"}` + "\n", ""},
		{"missing percentage value", "/percentage", `{"percentage":10}`, nil, 400, `{"error":"value is required"}` + "\n", ""},
		{"missing percentage", "/percentage", `{"value":100}`, nil, 400, `{"error":"percentage is required"}` + "\n", ""},
		{"wrong value type", "/addition", `{"operands":[1,"two"]}`, nil, 400, `{"error":"invalid JSON"}` + "\n", ""},
		{"null operand", "/addition", `{"operands":[1,null]}`, nil, 400, `{"error":"operands must contain only numbers"}` + "\n", ""},
		{"unknown field", "/square-root", `{"radicand":25,"extra":1}`, nil, 400, `{"error":"invalid JSON"}` + "\n", ""},
		{"one operand", "/multiplication", `{"operands":[2]}`, domain.ErrInsufficientOperands, 400, `{"error":"at least two operands are required"}` + "\n", "multiplication"},
		{"empty operands", "/addition", `{"operands":[]}`, domain.ErrInsufficientOperands, 400, `{"error":"at least two operands are required"}` + "\n", "addition"},
		{"division by zero service error", "/division", `{"operands":[10,2]}`, domain.ErrDivisionByZero, 400, `{"error":"division by zero"}` + "\n", "division"},
		{"negative square root service error", "/square-root", `{"radicand":4}`, domain.ErrNegativeRadicand, 400, `{"error":"square root of a negative number has no real result"}` + "\n", "square-root"},
		{"no real exponentiation result", "/exponentiation", `{"base":2,"exponent":2}`, domain.ErrNoRealResult, 400, `{"error":"calculation has no real-number result"}` + "\n", "exponentiation"},
		{"non-finite input service error", "/percentage", `{"value":100,"percentage":10}`, domain.ErrNonFiniteInput, 400, `{"error":"values must be finite real numbers"}` + "\n", "percentage"},
		{"non-representable service result", "/exponentiation", `{"base":2,"exponent":2}`, domain.ErrNonRepresentableResult, 400, `{"error":"result cannot be represented as a finite number"}` + "\n", "exponentiation"},
		{"trailing JSON", "/percentage", `{"value":100,"percentage":10}{}`, nil, 400, `{"error":"invalid JSON: request body must contain one JSON object"}` + "\n", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calculator := &stubCalculator{result: 42, err: test.serviceErr}
			router := NewRouter(calculator)
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Body.String() != test.wantBody {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantBody)
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", got)
			}
			if calculator.called != test.wantCall {
				t.Fatalf("service call = %q, want %q", calculator.called, test.wantCall)
			}
		})
	}
}

func TestUnexpectedServiceError(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	response := httptest.NewRecorder()

	handleJSON(
		response,
		request,
		func(struct{}) error { return nil },
		func(struct{}) (float64, error) { return 0, errors.New("unexpected service error") },
	)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if got := response.Body.String(); got != `{"error":"unexpected service error"}`+"\n" {
		t.Fatalf("body = %q", got)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	calculator := &stubCalculator{}
	router := NewRouter(calculator)
	request := httptest.NewRequest(http.MethodGet, "/addition", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if got := response.Header().Get("Allow"); got != http.MethodPost {
		t.Fatalf("Allow = %q, want POST", got)
	}
	if got := response.Body.String(); got != `{"error":"method not allowed"}`+"\n" {
		t.Fatalf("body = %q", got)
	}
	if calculator.called != "" {
		t.Fatalf("service called for unsupported method: %q", calculator.called)
	}
}

type stubCalculator struct {
	result float64
	err    error
	called string
}

func (s *stubCalculator) Add([]float64) (float64, error) {
	return s.respond("addition")
}

func (s *stubCalculator) Multiply([]float64) (float64, error) {
	return s.respond("multiplication")
}

func (s *stubCalculator) Divide([]float64) (float64, error) {
	return s.respond("division")
}

func (s *stubCalculator) Exponentiate(float64, float64) (float64, error) {
	return s.respond("exponentiation")
}

func (s *stubCalculator) SquareRoot(float64) (float64, error) {
	return s.respond("square-root")
}

func (s *stubCalculator) Percentage(float64, float64) (float64, error) {
	return s.respond("percentage")
}

func (s *stubCalculator) respond(operation string) (float64, error) {
	s.called = operation
	return s.result, s.err
}
