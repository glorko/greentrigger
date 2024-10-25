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
	err = haproxyClient.BindGoService("go-service-backend", &services.ServiceProcess{
		Config: &services.ServiceConfig{
			Name: "go-service",
			URL:  "/go",
		},
		Port: 8080, // Hardcoded port for now
	})
	if err != nil {
		fmt.Printf("Error binding Go service to HAProxy: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Bound Go service to HAProxy")

	// Set up the root URL listener
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested path matches the configured URL
		if r.URL.Path == config.URL {
			// Start the Java service
			javaService := &services.ServiceProcess{
				Config: &services.ServiceConfig{
					Name: "java-service",
					URL:  config.URL,
				},
				Port: 8081, // Hardcoded port for now
			}
			err = javaService.Start()
			if err != nil {
				fmt.Printf("Error starting Java service: %v\n", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			fmt.Println("Started Java service")

			// Unbind the Go service from HAProxy
			err = haproxyClient.UnbindService("go-service-backend")
			if err != nil {
				fmt.Printf("Error unbinding Go service from HAProxy: %v\n", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			fmt.Println("Unbound Go service from HAProxy")

			// Bind the Java service to HAProxy
			err = haproxyClient.BindJavaService("java-service-backend", javaService)
			if err != nil {
				fmt.Printf("Error binding Java service to HAProxy: %v\n", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			fmt.Println("Bound Java service to HAProxy")

			// Redirect the request to the Java service
			http.Redirect(w, r, fmt.Sprintf("http://localhost:%d%s", javaService.Port, config.URL), http.StatusTemporaryRedirect)
		} else {
			// If the requested path doesn't match, return a 503 error
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		}
	})

	fmt.Println("Starting listener on :8000")
	http.ListenAndServe(":8000", nil)
}
