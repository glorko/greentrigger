package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// AddToHAProxy adds the Java service process to the HAProxy backend dynamically.
func AddToHAProxy(process *ServiceProcess) error {
	apiURL := "http://localhost:5555/v3/services/haproxy"

	// Step 1: Retrieve the current HAProxy configuration version.
	cfgVer, err := getCurrentConfigVersion(apiURL)
	if err != nil {
		return fmt.Errorf("failed to retrieve configuration version: %v", err)
	}

	// Step 2: Start a transaction for modifying the configuration.
	transactionID, err := startTransaction(apiURL, cfgVer)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// Step 3: Check if the backend exists and create it if not.
	backendName := "service-backend"
	err = createBackend(apiURL, backendName, transactionID)
	if err != nil {
		rollbackTransaction(apiURL, transactionID)
		return fmt.Errorf("failed to create backend: %v", err)
	}

	// Step 4: Add the Java service as a server to the backend.
	err = addServer(apiURL, backendName, process, transactionID)
	if err != nil {
		rollbackTransaction(apiURL, transactionID)
		return fmt.Errorf("failed to add server to backend: %v", err)
	}

	// Step 5: Commit the transaction.
	err = commitTransaction(apiURL, transactionID)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	fmt.Printf("Server %s added to backend %s successfully\n", process.Config.Name, backendName)
	return nil
}

// getCurrentConfigVersion retrieves the current HAProxy configuration version as an integer.
func getCurrentConfigVersion(apiURL string) (int64, error) {
	url := fmt.Sprintf("%s/configuration/version", apiURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create version request: %v", err)
	}

	req.SetBasicAuth("admin", "mypassword")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve version: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to retrieve version, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	// Read the version as a string and clean it
	versionStr, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read version response: %v", err)
	}

	// Trim any newline or whitespace characters
	cleanedVersionStr := strings.TrimSpace(string(versionStr))

	// Convert the cleaned version string to an integer
	version, err := strconv.ParseInt(cleanedVersionStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse version as integer: %v", err)
	}

	return version, nil
}

// startTransaction starts a new HAProxy configuration transaction.
func startTransaction(apiURL string, version int64) (string, error) {
	// Prepare the URL with proper query encoding
	u, err := url.Parse(fmt.Sprintf("%s/transactions", apiURL))
	if err != nil {
		return "", fmt.Errorf("failed to parse API URL: %v", err)
	}

	query := u.Query()
	query.Set("version", strconv.FormatInt(version, 10))
	u.RawQuery = query.Encode()

	// Create the POST request for starting a transaction
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction request: %v", err)
	}

	req.SetBasicAuth("admin", "mypassword")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to start transaction, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	// Parse the transaction ID from the response
	var transactionData struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&transactionData); err != nil {
		return "", fmt.Errorf("failed to parse transaction response: %v", err)
	}

	return transactionData.ID, nil
}

func commitTransaction(apiURL, transactionID string) error {
	url := fmt.Sprintf("%s/transactions/%s", apiURL, transactionID)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create commit transaction request: %v", err)
	}

	req.SetBasicAuth("admin", "mypassword")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to commit transaction, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func rollbackTransaction(apiURL, transactionID string) error {
	url := fmt.Sprintf("%s/transactions/%s", apiURL, transactionID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create rollback transaction request: %v", err)
	}

	req.SetBasicAuth("admin", "mypassword")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to rollback transaction, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func createBackend(apiURL, backendName, transactionID string) error {
	url := fmt.Sprintf("%s/configuration/backends?transaction_id=%s", apiURL, transactionID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth("admin", "mypassword")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check if backend exists: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		backendData := map[string]interface{}{
			"name": backendName,
			"mode": "http",
			"balance": map[string]string{
				"algorithm": "roundrobin",
			},
		}
		backendPayload, _ := json.Marshal(backendData)

		createURL := fmt.Sprintf("%s/configuration/backends?transaction_id=%s", apiURL, transactionID)
		req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(backendPayload))
		if err != nil {
			return fmt.Errorf("failed to create request to create backend: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("admin", "mypassword")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to create backend: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to create backend, status code: %d, response: %s", resp.StatusCode, string(body))
		}

		fmt.Printf("Backend %s created successfully\n", backendName)
	} else if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to check backend, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func addServer(apiURL, backendName string, process *ServiceProcess, transactionID string) error {
	serverData := map[string]interface{}{
		"name":    process.Config.Name,
		"address": "localhost",
		"port":    process.Port,
	}
	serverPayload, _ := json.Marshal(serverData)

	createURL := fmt.Sprintf("%s/configuration/backends/%s/servers?transaction_id=%s", apiURL, backendName, transactionID)
	req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(serverPayload))
	if err != nil {
		return fmt.Errorf("failed to create request to add server: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "mypassword")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add server to backend: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add server to backend, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}
