package services

import (
	"fmt"
	"net/http"
	"strings"
)

func AddToHAProxy(process *ServiceProcess) error {
	haproxyURL := "http://localhost:8404/haproxy?stats" // Access stats API on port 8404

	// Command to add the real service to the pre-configured backend (service-backend)
	serverCommand := fmt.Sprintf("add server service-backend/%s 127.0.0.1:%d check", process.Config.Name, process.Port)
	addServerReq, err := http.NewRequest("POST", haproxyURL, strings.NewReader(serverCommand))
	if err != nil {
		return fmt.Errorf("failed to create request to add server: %v", err)
	}

	addServerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	serverResp, err := client.Do(addServerReq)
	if err != nil {
		return fmt.Errorf("failed to add server to backend: %v", err)
	}
	defer serverResp.Body.Close()

	if serverResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add server to backend, status code: %d", serverResp.StatusCode)
	}

	fmt.Printf("Server %s added to service-backend on port %d successfully\n", process.Config.Name, process.Port)

	return nil
}
