package services

import (
	"fmt"
	"net/http"
)

func StartServer(service ServiceConfig) {
	http.HandleFunc(service.URL, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s!", service.Name)
	})

	fmt.Printf("Starting server %s on %s...\n", service.Name, service.URL)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server %s: %v\n", service.Name, err)
	}
}
