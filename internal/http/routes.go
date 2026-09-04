package http

import (
	stdhttp "net/http"

	"simple_calculator/internal/service"
)

func NewRouter(calculator *service.Calculator) stdhttp.Handler {
	handler := NewHandler(calculator)
	mux := stdhttp.NewServeMux()
	mux.HandleFunc("/addition", postOnly(handler.addition))
	mux.HandleFunc("/multiplication", postOnly(handler.multiplication))
	mux.HandleFunc("/division", postOnly(handler.division))
	mux.HandleFunc("/exponentiation", postOnly(handler.exponentiation))
	mux.HandleFunc("/square-root", postOnly(handler.squareRoot))
	mux.HandleFunc("/percentage", postOnly(handler.percentage))
	return mux
}

func postOnly(next stdhttp.HandlerFunc) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.Method != stdhttp.MethodPost {
			w.Header().Set("Allow", stdhttp.MethodPost)
			writeError(w, stdhttp.StatusMethodNotAllowed, "method not allowed")
			return
		}
		next(w, r)
	}
}
