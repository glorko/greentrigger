package haproxy

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
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
	//client.SetDebug(true)
	client.SetHeader("Content-Type", "application/json")
	client.SetDisableWarn(true)
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

func (c *HAProxyConfigurationManager) StartTransaction(version int64) (string, error) {
	resp, err := c.client.R().
		SetQueryParam("version", strconv.FormatInt(version, 10)).
		Post("/transactions")
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %v", err)
	}

	if resp.StatusCode() != 201 {
		return "", fmt.Errorf("failed to start transaction, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	var transaction struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &transaction); err != nil {
		return "", fmt.Errorf("failed to parse transaction ID: %v", err)
	}

	return transaction.ID, nil
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
func (c *HAProxyConfigurationManager) DeleteServer(backendName, serverName, transactionID string) error {
	resp, err := c.client.R().
		SetQueryParam("transaction_id", transactionID).
		Delete(fmt.Sprintf("/configuration/backends/%s/servers/%s", backendName, serverName))

	if err != nil {
		return fmt.Errorf("failed to delete server %s from backend %s: %v", serverName, backendName, err)
	}

	switch resp.StatusCode() {
	case 204, 202: // Accept both immediate success and accepted for reload
		log.Printf("[INFO] Server %s successfully deleted from backend %s", serverName, backendName)
		return nil
	case 404:
		var apiErr struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(resp.Body(), &apiErr); err != nil {
			return fmt.Errorf("failed to parse error response: %v", err)
		}
		log.Printf("[INFO] Server or backend not found: %s", apiErr.Message)
		return nil
	case 400:
		return fmt.Errorf("API error deleting server %s from backend %s: %s",
			serverName, backendName, resp.String())
	default:
		return fmt.Errorf("unexpected status %d deleting server %s from backend %s: %s",
			resp.StatusCode(), serverName, backendName, resp.String())
	}
}
