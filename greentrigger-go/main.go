package main

import (
	"fmt"
	"greentrigger/services"
	"greentrigger/services/haproxy"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./GreenTrigger <path-to-config>")
		os.Exit(1)
	}
	configPath := os.Args[1]

	// Read the config
	config, err := services.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create an instance of the HAProxyClient
	haproxyClient := haproxy.NewHAProxyClient()

	// Bind the Go service to HAProxy
	err = haproxyClient.BindGoService("service-backend", config.Services[0])
	if err != nil {
		fmt.Printf("Error binding Go service to HAProxy: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Bound Go service to HAProxy")

	// Set up the root URL listener
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested path matches the configured URL
		if r.URL.Path == config.Services[0].URL {
			// Start the Java service
			javaService, err := services.StartServer(&config.Services[0])
			if err != nil {
				fmt.Printf("Error starting Java service: %v\n", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			fmt.Println("Started Java service")

			// Unbind the Go service from HAProxy
			err = haproxyClient.UnbindService("go-service-backend", "go-service")
			if err != nil {
				fmt.Printf("Error unbinding Go service from HAProxy: %v\n", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			fmt.Println("Unbound Go service from HAProxy")

			// Bind the Java service to HAProxy
			//err = haproxyClient.BindService("java-service-backend", "java-service", "localhost", javaService.Port)
			//if err != nil {
			//	fmt.Printf("Error binding Java service to HAProxy: %v\n", err)
			//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			//	return
			//}
			fmt.Println("Bound Java service to HAProxy")

			// Redirect the request to the Java service
			http.Redirect(w, r, fmt.Sprintf("http://localhost:%d%s", javaService.Port, config.Services[0].URL), http.StatusTemporaryRedirect)
		} else {
			// If the requested path doesn't match, return a 503 error
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		}
	})

	fmt.Println("Starting listener on :8000")
	http.ListenAndServe(":8000", nil)
}
