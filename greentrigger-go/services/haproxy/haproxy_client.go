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
	return c.transactionMiddleware(func(transactionID string, version int64) error {
		// Create the backend if it doesn't exist
		err := c.configManager.CreateBackend(backendName, transactionID)
		if err != nil {
			return fmt.Errorf("failed to create backend: %v", err)
		}

		// Delete both services
		for _, serviceName := range []string{"go-service", targetService.Name} {
			err = c.configManager.DeleteServer(backendName, serviceName, transactionID, version)
			if err != nil {
				log.Printf("[WARN] Could not delete service %s: %v", serviceName, err)
			}
		}

		// Add the Go service with hardcoded config
		err = c.configManager.AddServer(backendName, map[string]interface{}{
			"name":    "go-service",
			"address": "localhost",
			"port":    8000, // Hardcoded here instead of main
		}, transactionID, version)
		if err != nil {
			return fmt.Errorf("failed to add server to backend: %v", err)
		}

		return nil
	})()
}

// UnbindService removes the service from the HAProxy configuration.
func (c *HAProxyClient) UnbindService(backendName, serviceName string) error {
	return c.transactionMiddleware(func(transactionID string, version int64) error {
		// Replace the service with a placeholder
		err := c.configManager.ReplaceServer(backendName, serviceName, transactionID, version, map[string]interface{}{
			"name":    serviceName,
			"address": "127.0.0.1",
			"port":    0,
		})
		if err != nil {
			return fmt.Errorf("failed to remove service from backend: %v", err)
		}

		return nil
	})()
}

// BindService replaces the existing service with the new service in the HAProxy configuration.
func (c *HAProxyClient) BindService(backendName, serviceName, serviceAddress string, servicePort int, version int64) error {
	return c.transactionMiddleware(func(transactionID string, version int64) error {
		// Replace the existing service with the new service
		err := c.configManager.ReplaceServer(backendName, serviceName, transactionID, version, map[string]interface{}{
			"name":    serviceName,
			"address": serviceAddress,
			"port":    servicePort,
		})
		if err != nil {
			return fmt.Errorf("failed to replace existing service with new service: %v", err)
		}

		return nil
	})()
}
