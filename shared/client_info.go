package shared

import (
	"os"
	"runtime"
	"time"
)

// ClientInfo holds information about a connected client.
type ClientInfo struct {
	ID       string `json:"id"`
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	OSInfo   string `json:"os_info"`
}

// BuildClientInfo creates a ClientInfo with system details.
func BuildClientInfo(id string) ClientInfo {
	hostname, _ := os.Hostname()

	osInfo := buildOSInfo(hostname)

	return ClientInfo{
		ID:       id,
		IP:       "unknown",
		Hostname: hostname,
		OSInfo:   osInfo,
	}
}

// buildOSInfo constructs an OS info string from runtime details.
func buildOSInfo(hostname string) string {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	version := runtime.Version()

	return osName + " " + hostname + " " + version + " " + arch
}

// ToPayload converts ClientInfo to a map for JSON serialization.
func (c ClientInfo) ToPayload() map[string]interface{} {
	return map[string]interface{}{
		"id":       c.ID,
		"ip":       c.IP,
		"hostname": c.Hostname,
		"os_info":  c.OSInfo,
	}
}

// ClientInfoFromPayload creates a ClientInfo from a JSON payload map.
func ClientInfoFromPayload(p map[string]interface{}) ClientInfo {
	info := ClientInfo{}
	if v, ok := p["id"]; ok {
		info.ID, _ = v.(string)
	}
	if v, ok := p["ip"]; ok {
		info.IP, _ = v.(string)
	}
	if v, ok := p["hostname"]; ok {
		info.Hostname, _ = v.(string)
	}
	if v, ok := p["os_info"]; ok {
		info.OSInfo, _ = v.(string)
	}
	return info
}

// GenerateClientID creates a unique client ID from hostname and UUID.
func GenerateClientID() string {
	hostname, _ := os.Hostname()
	return hostname + "-" + GenerateID()[:8]
}

// CalculateBackoff computes exponential backoff duration capped at 30 seconds.
func CalculateBackoff(attempt int) time.Duration {
	backoff := time.Duration(1<<uint(attempt)) * time.Second
	if backoff > 30*time.Second {
		backoff = 30 * time.Second
	}
	return backoff
}
