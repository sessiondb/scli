package main

import (
	"fmt"
	"net/http"
	"time"
)

// runStatus checks if the SessionDB server is reachable at baseURL (default http://localhost:8080).
func runStatus(baseURL string) error {
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		fmt.Println("Status: unreachable")
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Status: running")
		return nil
	}
	fmt.Printf("Status: HTTP %d\n", resp.StatusCode)
	return nil
}
