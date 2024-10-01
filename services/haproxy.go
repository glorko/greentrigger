package services

import (
	"fmt"
	"net/http"
	"strings"
)

func AddToHAProxy(service ServiceConfig) error {
	haproxyURL := "http://localhost:8080/haproxy?stats" // Replace with actual HAProxy API URL

	data := fmt.Sprintf("add server backend/%s 127.0.0.1:8080", service.Name)
	req, err := http.NewRequest("POST", haproxyURL, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add server to HAProxy: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add server to HAProxy, status code: %d", resp.StatusCode)
	}

	fmt.Printf("Service %s added to HAProxy\n", service.Name)
	return nil
}
