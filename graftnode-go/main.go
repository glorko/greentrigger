package main

import (
	"fmt"
	"graftnode/services"
	"graftnode/services/haproxy"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./GraftNode <path-to-config>")
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
		if r.URL.Path != config.Services[0].URL {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		// a) Start Java server and rebind in HAProxy
		javaService, err := services.StartServer(&config.Services[0])
		if err != nil {
			fmt.Printf("Error starting service: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		fmt.Printf("Started %s service\n", config.Services[0].Name)

		err = haproxyClient.ReplaceGoStub("service-backend", javaService)
		if err != nil {
			fmt.Printf("Error replacing service in HAProxy: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		fmt.Printf("Replaced go-service with %s service in HAProxy\n", config.Services[0].Name)

		// b & c) Proxy this request to Java server and return response
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("localhost:%d", javaService.Port),
		})
		proxy.ServeHTTP(w, r)

		// d) Exit the program as go listener is no longer needed
		//go func() {
		//	fmt.Println("Shutting down go listener as service is started")
		//	os.Exit(0)
		//}()
	})
	fmt.Println("Starting listener on :8000")
	http.ListenAndServe(":8000", nil)
}
