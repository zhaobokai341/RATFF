package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"RATFF/shared"
)

// apiGet performs an authenticated GET request to the API.
func apiGet(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", getAPIBaseURL()+"/api"+path, nil)
	if err != nil {
		return nil, err
	}

	if jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+jwtToken)
	}

	return http.DefaultClient.Do(req)
}

// apiPost performs an authenticated POST request to the API.
func apiPost(path string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", getAPIBaseURL()+"/api"+path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+jwtToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// fetchClients retrieves the list of connected clients from the API.
func fetchClients() ([]shared.ClientInfo, error) {
	resp, err := apiGet("/clients")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Clients []shared.ClientInfo `json:"clients"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Clients, nil
}

// postCommand sends a command to the server API.
func postCommand(payload map[string]interface{}) error {
	return apiPost("/command", payload)
}
