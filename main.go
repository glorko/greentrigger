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

	// Start the services
	for _, service := range config.Services {
		go func(service services.ServiceConfig) {
			// Start the server and get the process details
			process, err := services.StartServer(&service)
			if err != nil {
				fmt.Printf("Failed to start service %s: %v\n", service.Name, err)
				return // Skip adding to HAProxy if the service didn't start
			}

			// Add the service to HAProxy using the process (which includes the port)
			err = services.AddToHAProxy(process)
			if err != nil {
				fmt.Printf("Error adding service %s to HAProxy: %v\n", service.Name, err)
				// Optionally, stop the process if adding to HAProxy fails
				process.Command.Process.Kill()
			}
		}(service) // Pass `service` explicitly to avoid issues with goroutines
	}

	// Block the main function from exiting (prevent deadlock)
	select {}
}
