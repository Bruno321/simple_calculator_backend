package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simple_calculator/internal/service"
)

func TestCalculatorEndpoints(t *testing.T) {
	router := NewRouter(service.NewCalculator())
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{"addition", "/addition", `{"operands":[10,4,2]}`, 200, `{"result":16}` + "\n"},
		{"multiplication", "/multiplication", `{"operands":[1.5,4]}`, 200, `{"result":6}` + "\n"},
		{"division", "/division", `{"operands":[10,4,2]}`, 200, `{"result":1.25}` + "\n"},
		{"exponentiation", "/exponentiation", `{"base":10,"exponent":2}`, 200, `{"result":100}` + "\n"},
		{"square root", "/square-root", `{"radicand":25}`, 200, `{"result":5}` + "\n"},
		{"percentage", "/percentage", `{"value":100,"percentage":10}`, 200, `{"result":10}` + "\n"},
		{"malformed JSON", "/addition", `{"operands":[1,2]`, 400, `{"error":"invalid JSON"}` + "\n"},
		{"missing operands", "/addition", `{}`, 400, `{"error":"operands is required"}` + "\n"},
		{"missing base", "/exponentiation", `{"exponent":2}`, 400, `{"error":"base is required"}` + "\n"},
		{"missing exponent", "/exponentiation", `{"base":10}`, 400, `{"error":"exponent is required"}` + "\n"},
		{"missing radicand", "/square-root", `{}`, 400, `{"error":"radicand is required"}` + "\n"},
		{"missing percentage value", "/percentage", `{"percentage":10}`, 400, `{"error":"value is required"}` + "\n"},
		{"missing percentage", "/percentage", `{"value":100}`, 400, `{"error":"percentage is required"}` + "\n"},
		{"wrong value type", "/addition", `{"operands":[1,"two"]}`, 400, `{"error":"invalid JSON"}` + "\n"},
		{"null operand", "/addition", `{"operands":[1,null]}`, 400, `{"error":"operands must contain only numbers"}` + "\n"},
		{"unknown field", "/square-root", `{"radicand":25,"extra":1}`, 400, `{"error":"invalid JSON"}` + "\n"},
		{"insufficient operands", "/multiplication", `{"operands":[2]}`, 400, `{"error":"at least two operands are required"}` + "\n"},
		{"division by zero", "/division", `{"operands":[10,0]}`, 400, `{"error":"division by zero"}` + "\n"},
		{"negative square root", "/square-root", `{"radicand":-1}`, 400, `{"error":"square root of a negative number has no real result"}` + "\n"},
		{"NaN exponentiation", "/exponentiation", `{"base":-2,"exponent":0.5}`, 400, `{"error":"calculation has no real-number result"}` + "\n"},
		{"infinite exponentiation", "/exponentiation", `{"base":10,"exponent":309}`, 400, `{"error":"result cannot be represented as a finite number"}` + "\n"},
		{"trailing JSON", "/percentage", `{"value":100,"percentage":10}{}`, 400, `{"error":"invalid JSON: request body must contain one JSON object"}` + "\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
	router := NewRouter(service.NewCalculator())
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
}
