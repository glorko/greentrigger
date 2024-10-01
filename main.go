package main

import (
  "fmt"
  "greentrigger/services"
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

	// Start listening on the configured service URL
	for _, service := range config.Services {
		go services.StartServer(service) // Start server in a separate goroutine
	}

	// Add the service to HAProxy via the HTTP API
	for _, service := range config.Services {
		err := services.AddToHAProxy(service)
		if err != nil {
			fmt.Printf("Error adding service %s to HAProxy: %v\n", service.Name, err)
			os.Exit(1)
		}
	}

	// Block the main function from exiting
	select {}
}
