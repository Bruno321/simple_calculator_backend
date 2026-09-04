package main

import (
	"log"
	"net/http"

	calculatorhttp "simple_calculator/internal/http"
	"simple_calculator/internal/service"
)

func main() {
	router := calculatorhttp.NewRouter(service.NewCalculator())
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Printf("calculator API listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
