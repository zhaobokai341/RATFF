package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"RATFF/shared"
)

// Client handles API requests with authentication.
type Client struct {
	baseURL string
	token   string
}

// NewClient creates a new API client.
func NewClient(baseURL, token string) *Client {
	return &Client{baseURL: baseURL, token: token}
}

// Get performs an authenticated GET request to the API.
func (c *Client) Get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api"+path, nil)
	if err != nil {
		return nil, err
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return http.DefaultClient.Do(req)
}

// Post performs an authenticated POST request to the API.
func (c *Client) Post(path string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api"+path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if jsonErr := json.NewDecoder(resp.Body).Decode(&errResp); jsonErr == nil && errResp.Error != "" {
			return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// FetchClients retrieves the list of connected clients from the API.
func (c *Client) FetchClients() ([]shared.ClientInfo, error) {
	resp, err := c.Get("/clients")
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

// PostCommand sends a command to the server API.
func (c *Client) PostCommand(payload map[string]interface{}) error {
	return c.Post("/command", payload)
}
