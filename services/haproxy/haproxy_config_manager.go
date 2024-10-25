package haproxy

import (
	"fmt"
	"strconv"
)

// HAProxyConfigurationManager provides low-level functions for managing the HAProxy configuration.
type HAProxyConfigurationManager struct {
	client *resty.Client
}

// NewHAProxyConfigurationManager creates a new instance of HAProxyConfigurationManager.
func NewHAProxyConfigurationManager(apiURL, username, password string) *HAProxyConfigurationManager {
	client := resty.New()
	client.SetBaseURL(apiURL)
	client.SetBasicAuth(username, password)
	client.SetHeader("Content-Type", "application/json")

	return &HAProxyConfigurationManager{
		client: client,
	}
}

// GetCurrentConfigVersion retrieves the current HAProxy configuration version as an integer.
func (c *HAProxyConfigurationManager) GetCurrentConfigVersion() (int64, error) {
	resp, err := c.client.R().Get("/configuration/version")
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve version: %v", err)
	}

	if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("failed to retrieve version, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	version, err := strconv.ParseInt(resp.String(), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse version as integer: %v", err)
	}

	return version, nil
}

// StartTransaction starts a new HAProxy configuration transaction.
func (c *HAProxyConfigurationManager) StartTransaction(version int64) (string, error) {
	resp, err := c.client.R().SetQueryParam("version", strconv.FormatInt(version, 10)).Post("/transactions")
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %v", err)
	}

	if resp.StatusCode() != 201 {
		return "", fmt.Errorf("failed to start transaction, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	return resp.String(), nil
}

// CommitTransaction commits the specified HAProxy configuration transaction.
func (c *HAProxyConfigurationManager) CommitTransaction(transactionID string) error {
	resp, err := c.client.R().Put(fmt.Sprintf("/transactions/%s", transactionID))
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	if resp.StatusCode() != 202 {
		return fmt.Errorf("failed to commit transaction, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// RollbackTransaction rolls back the specified HAProxy configuration transaction.
func (c *HAProxyConfigurationManager) RollbackTransaction(transactionID string) error {
	resp, err := c.client.R().Delete(fmt.Sprintf("/transactions/%s", transactionID))
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %v", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to rollback transaction, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// CreateBackend creates a new backend in the HAProxy configuration.
func (c *HAProxyConfigurationManager) CreateBackend(backendName, transactionID string) error {
	resp, err := c.client.R().SetQueryParam("transaction_id", transactionID).Get("/configuration/backends")
	if err != nil {
		return fmt.Errorf("failed to check if backend exists: %v", err)
	}

	if resp.StatusCode() == 404 {
		backendData := map[string]interface{}{
			"name": backendName,
			"mode": "http",
			"balance": map[string]string{
				"algorithm": "roundrobin",
			},
		}
		_, err = c.client.R().
			SetQueryParam("transaction_id", transactionID).
			SetBody(backendData).
			Post("/configuration/backends")
		if err != nil {
			return fmt.Errorf("failed to create backend: %v", err)
		}

		fmt.Printf("Backend %s created successfully\n", backendName)
	} else if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to check backend, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// AddServer adds a new server to the specified backend in the HAProxy configuration.
func (c *HAProxyConfigurationManager) AddServer(backendName string, serverData map[string]interface{}, transactionID string) error {
	_, err := c.client.R().
		SetQueryParam("transaction_id", transactionID).
		SetBody(serverData).
		Post(fmt.Sprintf("/configuration/backends/%s/servers", backendName))
	if err != nil {
		return fmt.Errorf("failed to add server to backend: %v", err)
	}

	return nil
}

// ReplaceServer replaces the configuration of a server in the specified backend.
func (c *HAProxyConfigurationManager) ReplaceServer(backendName, serverName, transactionID string, serverData map[string]interface{}) error {
	_, err := c.client.R().
		SetQueryParam("backend", backendName).
		SetQueryParam("transaction_id", transactionID).
		SetBody(serverData).
		Put(fmt.Sprintf("/services/haproxy/configuration/servers/%s", serverName))
	if err != nil {
		return fmt.Errorf("failed to replace server in backend: %v", err)
	}

	return nil
}
