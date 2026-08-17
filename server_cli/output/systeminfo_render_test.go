package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintHostInfo(t *testing.T) {
	payload := map[string]interface{}{
		"host": map[string]interface{}{
			"hostname":         "test-host",
			"os":               "linux",
			"platform":         "ubuntu",
			"platform_family":  "debian",
			"platform_version": "22.04",
			"kernel_version":   "5.15.0",
			"kernel_arch":      "x86_64",
			"uptime":           86400.0,
			"boot_time":        "2024-01-01",
			"procs":            100.0,
		},
	}
	printHostInfo(payload)
}

func TestPrintCPUInfo(t *testing.T) {
	payload := map[string]interface{}{
		"cpu": map[string]interface{}{
			"cpus": []interface{}{
				map[string]interface{}{
					"model_name": "Intel Core i7",
					"cores":      8.0,
					"mhz":        3500.0,
					"cache_size": "12 MB",
					"vendor_id":  "GenuineIntel",
				},
			},
		},
	}
	printCPUInfo(payload)
}

func TestPrintCPUInfoInvalid(t *testing.T) {
	payload := map[string]interface{}{
		"cpu": map[string]interface{}{
			"cpus": "invalid",
		},
	}
	printCPUInfo(payload)
}

func TestPrintMemoryInfo(t *testing.T) {
	payload := map[string]interface{}{
		"memory": map[string]interface{}{
			"total":        16777216.0,
			"used":         8388608.0,
			"available":    8388608.0,
			"free":         4194304.0,
			"used_percent": 50.0,
			"active":       2097152.0,
			"inactive":     1048576.0,
			"buffers":      524288.0,
			"cached":       1048576.0,
		},
	}
	printMemoryInfo(payload)
}

func TestPrintSwapMemoryInfo(t *testing.T) {
	payload := map[string]interface{}{
		"swap_memory": map[string]interface{}{
			"total":        4294967296.0,
			"used":         1073741824.0,
			"free":         3221225472.0,
			"used_percent": 25.0,
			"sin":          0.0,
			"sout":         0.0,
		},
	}
	printSwapMemoryInfo(payload)
}

func TestPrintPartitionInfo(t *testing.T) {
	payload := map[string]interface{}{
		"partition": map[string]interface{}{
			"partitions": []interface{}{
				map[string]interface{}{
					"device":       "/dev/sda1",
					"mountpoint":   "/",
					"fstype":       "ext4",
					"total":        107374182400.0,
					"used":         53687091200.0,
					"free":         53687091200.0,
					"used_percent": 50.0,
				},
			},
		},
	}
	printPartitionInfo(payload)
}

func TestPrintPartitionInfoWithError(t *testing.T) {
	payload := map[string]interface{}{
		"partition": map[string]interface{}{
			"partitions": []interface{}{
				map[string]interface{}{
					"device":      "/dev/sdb1",
					"mountpoint":  "/mnt/data",
					"fstype":      "ext4",
					"usage_error": "permission denied",
				},
			},
		},
	}
	printPartitionInfo(payload)
}

func TestPrintPartitionInfoInvalid(t *testing.T) {
	payload := map[string]interface{}{
		"partition": map[string]interface{}{
			"partitions": "invalid",
		},
	}
	printPartitionInfo(payload)
}

func TestPrintIODiskInfo(t *testing.T) {
	payload := map[string]interface{}{
		"io_disk": map[string]interface{}{
			"sda": map[string]interface{}{
				"read_count":  1000.0,
				"write_count": 500.0,
				"read_bytes":  1048576.0,
				"write_bytes": 524288.0,
				"read_time":   100.0,
				"write_time":  50.0,
			},
		},
	}
	printIODiskInfo(payload)
}

func TestPrintInterfacesInfo(t *testing.T) {
	payload := map[string]interface{}{
		"interfaces": map[string]interface{}{
			"interfaces": []interface{}{
				map[string]interface{}{
					"name":  "eth0",
					"mtu":   1500.0,
					"flags": "up,broadcast,multicast",
					"addresses": []interface{}{
						"192.168.1.100",
						"fe80::1",
					},
				},
			},
		},
	}
	printInterfacesInfo(payload)
}

func TestPrintInterfacesInfoInvalid(t *testing.T) {
	payload := map[string]interface{}{
		"interfaces": map[string]interface{}{
			"interfaces": "invalid",
		},
	}
	printInterfacesInfo(payload)
}

func TestPrintIONetworkInfo(t *testing.T) {
	payload := map[string]interface{}{
		"io_network": map[string]interface{}{
			"io_counters": []interface{}{
				map[string]interface{}{
					"name":         "eth0",
					"bytes_sent":   1048576.0,
					"bytes_recv":   2097152.0,
					"packets_sent": 1000.0,
					"packets_recv": 2000.0,
					"errin":        0.0,
					"errout":       0.0,
					"dropin":       0.0,
					"dropout":      0.0,
				},
			},
		},
	}
	printIONetworkInfo(payload)
}

func TestPrintIONetworkInfoInvalid(t *testing.T) {
	payload := map[string]interface{}{
		"io_network": map[string]interface{}{
			"io_counters": "invalid",
		},
	}
	printIONetworkInfo(payload)
}

func TestPrintProcessesInfo(t *testing.T) {
	payload := map[string]interface{}{
		"processes": map[string]interface{}{
			"count": 100.0,
			"processes": []interface{}{
				map[string]interface{}{
					"pid":            1.0,
					"name":           "init",
					"status":         []interface{}{"running"},
					"cpu_percent":    0.5,
					"memory_percent": 1.2,
					"create_time":    "2024-01-01 00:00:00",
				},
			},
		},
	}
	printProcessesInfo(payload)
}

func TestPrintProcessesInfoInvalid(t *testing.T) {
	payload := map[string]interface{}{
		"processes": map[string]interface{}{
			"processes": "invalid",
		},
	}
	printProcessesInfo(payload)
}

func TestFormatBytesValue(t *testing.T) {
	assert.Equal(t, "N/A", FormatBytesValue(nil))
	assert.Equal(t, "1.0 KB", FormatBytesValue(1024.0))
	assert.Equal(t, "1.0 MB", FormatBytesValue(1048576.0))
}

func TestFormatBytesCompact(t *testing.T) {
	assert.Equal(t, "0 B", FormatBytesCompact(0))
	assert.Equal(t, "1.0 KB", FormatBytesCompact(1024.0))
	assert.Equal(t, "1.0 MB", FormatBytesCompact(1048576.0))
	assert.Equal(t, "1.0 GB", FormatBytesCompact(1073741824.0))
	assert.Equal(t, "1.0 TB", FormatBytesCompact(1099511627776.0))
}

func TestFormatUptimeValue(t *testing.T) {
	assert.Equal(t, "0s", FormatUptimeValue(0))
	assert.Equal(t, "1d 0s", FormatUptimeValue(uint64(86400)))
	assert.Equal(t, "1h 30m 45s", FormatUptimeValue(uint64(5445)))
	assert.Equal(t, "2d 3h 4m 5s", FormatUptimeValue(uint64(183845)))
}

func TestToFloat64(t *testing.T) {
	assert.Equal(t, 0.0, ToFloat64(nil))
	assert.Equal(t, 3.14, ToFloat64(3.14))
	assert.Equal(t, 0.0, ToFloat64("invalid"))
}

func TestToUint64(t *testing.T) {
	assert.Equal(t, uint64(0), ToUint64(nil))
	assert.Equal(t, uint64(100), ToUint64(100.0))
	assert.Equal(t, uint64(200), ToUint64(uint64(200)))
	assert.Equal(t, uint64(0), ToUint64("invalid"))
}

func TestToString(t *testing.T) {
	assert.Equal(t, "", ToString(nil))
	assert.Equal(t, "hello", ToString("hello"))
	assert.Equal(t, "123", ToString(123))
}

func TestPrintKeyValue(t *testing.T) {
	printKeyValue("Test Key", "Test Value")
}
