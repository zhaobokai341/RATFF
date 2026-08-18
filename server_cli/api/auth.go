package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LoginToAPI logs into server_api and returns a JWT token.
func LoginToAPI(apiBaseURL, password string, t func(string) string) (string, error) {
	body := map[string]string{"password": password}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(apiBaseURL+"/verify", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", errors.New(t("invalid_path_password"))
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New(t("invalid_login_password"))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("server returned invalid response: %s", string(respBody))
	}

	return result.Token, nil
}
