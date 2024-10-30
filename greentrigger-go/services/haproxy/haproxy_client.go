package haproxy

import (
	"fmt"
	"greentrigger/services"
	"log"
)

const haproxyAPIURL = "http://localhost:5555/v3/services/haproxy"
const haproxyUsername = "admin"
const haproxyPassword = "mypassword"

// HAProxyClient provides a high-level interface for managing the HAProxy configuration.
type HAProxyClient struct {
	configManager         *HAProxyConfigurationManager
	transactionMiddleware TransactionMiddleware
}

// NewHAProxyClient creates a new instance of HAProxyClient.
func NewHAProxyClient() *HAProxyClient {
	configManager := NewHAProxyConfigurationManager(haproxyAPIURL, haproxyUsername, haproxyPassword)
	transactionMiddleware := NewTransactionMiddleware(configManager)

	return &HAProxyClient{
		configManager:         configManager,
		transactionMiddleware: transactionMiddleware,
	}
}

func (c *HAProxyClient) BindGoService(backendName string, targetService services.ServiceConfig) error {
	return c.transactionMiddleware(func(transactionID string) error {
		// Create the backend if it doesn't exist
		err := c.configManager.CreateBackend(backendName, transactionID)
		if err != nil {
			return fmt.Errorf("failed to create backend: %v", err)
		}

		// Delete both services
		for _, serviceName := range []string{"go-service", targetService.Name} {
			err = c.configManager.DeleteServer(backendName, serviceName, transactionID)
			if err != nil {
				log.Printf("[WARN] Could not delete service %s: %v", serviceName, err)
			}
		}

		// Add the Go service with hardcoded config
		err = c.configManager.AddServer(backendName, map[string]interface{}{
			"name":    "go-service",
			"address": "localhost",
			"port":    8000, // Hardcoded here instead of main
		}, transactionID)
		if err != nil {
			return fmt.Errorf("failed to add server to backend: %v", err)
		}

		return nil
	})()
}
func (c *HAProxyClient) ReplaceGoStub(backendName string, serviceProcess *services.ServiceProcess) error {
	return c.transactionMiddleware(func(transactionID string) error {
		// Delete go-service
		err := c.configManager.DeleteServer(backendName, "go-service", transactionID)
		if err != nil {
			return fmt.Errorf("failed to delete go-service: %v", err)
		}

		// Add the new service
		err = c.configManager.AddServer(backendName, map[string]interface{}{
			"name":    serviceProcess.Config.Name,
			"address": "localhost",
			"port":    serviceProcess.Port,
		}, transactionID)
		if err != nil {
			return fmt.Errorf("failed to add new service: %v", err)
		}

		return nil
	})()
}

func (c *HAProxyClient) BindService(backendName, serviceName, serviceAddress string, servicePort int) error {
	return c.transactionMiddleware(func(transactionID string) error {
		return c.configManager.AddServer(backendName, map[string]interface{}{
			"name":    serviceName,
			"address": serviceAddress,
			"port":    servicePort,
		}, transactionID)
	})()
}
