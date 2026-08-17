package output

import (
	"testing"

	"RATFF/shared"

	"github.com/stretchr/testify/assert"
)

func TestPrintClientTableEmpty(t *testing.T) {
	tFunc := func(key string) string { return key }
	tfFunc := func(key string, args ...interface{}) string { return key }
	PrintClientTable([]shared.ClientInfo{}, tFunc, tfFunc)
}

func TestPrintClientTableWithClients(t *testing.T) {
	tFunc := func(key string) string { return key }
	tfFunc := func(key string, args ...interface{}) string { return key }

	clients := []shared.ClientInfo{
		{ID: "client-1", IP: "127.0.0.1", Hostname: "host1", OSInfo: "linux"},
		{ID: "client-2", IP: "127.0.0.2", Hostname: "host2", OSInfo: "windows"},
	}
	PrintClientTable(clients, tFunc, tfFunc)
}

func TestFormatID(t *testing.T) {
	// Short ID
	assert.Equal(t, "short-id", FormatID("short-id"))

	// Exactly IDColumnWidth
	exactID := "12345678901234567890"
	assert.Equal(t, exactID, FormatID(exactID))

	// Longer than IDColumnWidth
	longID := "123456789012345678901234567890"
	result := FormatID(longID)
	assert.Equal(t, "12345678901234567...", result)
}

func TestPrintFileTableEmpty(t *testing.T) {
	tFunc := func(key string) string { return key }
	tfFunc := func(key string, args ...interface{}) string { return key }
	PrintFileTable("/home/user", []interface{}{}, tFunc, tfFunc)
}

func TestPrintFileTableWithFiles(t *testing.T) {
	tFunc := func(key string) string { return key }
	tfFunc := func(key string, args ...interface{}) string { return key }

	files := []interface{}{
		map[string]interface{}{
			"name":        "file1.txt",
			"type":        "file",
			"size":        1024.0,
			"mod_time":    1609459200.0,
			"permissions": "rw-r--r--",
			"hidden":      false,
			"link_target": "",
		},
		map[string]interface{}{
			"name":        "dir1",
			"type":        "directory",
			"size":        4096.0,
			"mod_time":    1609459200.0,
			"permissions": "rwxr-xr-x",
			"hidden":      false,
			"link_target": "",
		},
	}
	PrintFileTable("/home/user", files, tFunc, tfFunc)
}

func TestPrintFileTableWithHiddenAndSymlink(t *testing.T) {
	tFunc := func(key string) string { return key }
	tfFunc := func(key string, args ...interface{}) string { return key }

	files := []interface{}{
		map[string]interface{}{
			"name":        ".hidden",
			"type":        "file",
			"size":        512.0,
			"mod_time":    1609459200.0,
			"permissions": "rw-------",
			"hidden":      true,
			"link_target": "",
		},
		map[string]interface{}{
			"name":        "symlink",
			"type":        "symlink",
			"size":        0.0,
			"mod_time":    1609459200.0,
			"permissions": "lrwxrwxrwx",
			"hidden":      false,
			"link_target": "/target/path",
		},
	}
	PrintFileTable("/home/user", files, tFunc, tfFunc)
}

func TestPrintFileTableWithInvalidEntries(t *testing.T) {
	tFunc := func(key string) string { return key }
	tfFunc := func(key string, args ...interface{}) string { return key }

	files := []interface{}{
		"invalid",
		123,
		map[string]interface{}{
			"name":        "valid.txt",
			"type":        "file",
			"size":        100.0,
			"mod_time":    1609459200.0,
			"permissions": "rw-r--r--",
		},
	}
	PrintFileTable("/home/user", files, tFunc, tfFunc)
}

func TestGetFileTypeIcon(t *testing.T) {
	assert.Equal(t, "📁", getFileTypeIcon("directory"))
	assert.Equal(t, "🔗", getFileTypeIcon("symlink"))
	assert.Equal(t, "🔗", getFileTypeIcon("shortcut"))
	assert.Equal(t, "📄", getFileTypeIcon("file"))
	assert.Equal(t, "📄", getFileTypeIcon("unknown"))
}

func TestFormatFileSize(t *testing.T) {
	assert.Equal(t, "0 B", formatFileSize(0))
	assert.Equal(t, "100 B", formatFileSize(100))
	assert.Equal(t, "1.0 KB", formatFileSize(1024))
	assert.Equal(t, "1.0 MB", formatFileSize(1024*1024))
	assert.Equal(t, "1.0 GB", formatFileSize(1024*1024*1024))
	assert.Equal(t, "1.0 TB", formatFileSize(1024*1024*1024*1024))
}

func TestFormatModTime(t *testing.T) {
	assert.Equal(t, "N/A", formatModTime(0))
	assert.Equal(t, "2021-01-01 00:00:00", formatModTime(1609459200))
}
