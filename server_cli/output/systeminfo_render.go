package output

import (
	"fmt"
)

func printHostInfo(payload map[string]interface{}) {
	data := payload["host"].(map[string]interface{})

	fmt.Println(sysinfoSectionStyle.Render("=== Host ==="))
	printKeyValue("Hostname", data["hostname"])
	printKeyValue("OS", data["os"])
	printKeyValue("Platform", data["platform"])
	printKeyValue("Platform Family", data["platform_family"])
	printKeyValue("Platform Version", data["platform_version"])
	printKeyValue("Kernel Version", data["kernel_version"])
	printKeyValue("Kernel Arch", data["kernel_arch"])
	printKeyValue("Uptime", FormatUptimeValue(data["uptime"]))
	printKeyValue("Boot Time", data["boot_time"])
	printKeyValue("Processes", data["procs"])
	fmt.Println()
}

func printCPUInfo(payload map[string]interface{}) {
	data := payload["cpu"].(map[string]interface{})
	cpus, ok := data["cpus"].([]interface{})
	if !ok {
		return
	}

	fmt.Println(sysinfoSectionStyle.Render("=== CPU ==="))
	for i, cpuData := range cpus {
		if cpuMap, ok := cpuData.(map[string]interface{}); ok {
			fmt.Println(sysinfoTitleStyle.Render(fmt.Sprintf("  CPU %d:", i)))
			printKeyValue("Model", cpuMap["model_name"])
			printKeyValue("Cores", cpuMap["cores"])
			printKeyValue("MHz", cpuMap["mhz"])
			printKeyValue("Cache Size", cpuMap["cache_size"])
			printKeyValue("Vendor", cpuMap["vendor_id"])
		}
	}
	fmt.Println()
}

func printMemoryInfo(payload map[string]interface{}) {
	data := payload["memory"].(map[string]interface{})

	fmt.Println(sysinfoSectionStyle.Render("=== Memory ==="))
	printKeyValue("Total", FormatBytesValue(data["total"]))
	printKeyValue("Used", FormatBytesValue(data["used"]))
	printKeyValue("Available", FormatBytesValue(data["available"]))
	printKeyValue("Free", FormatBytesValue(data["free"]))
	printKeyValue("Used Percent", fmt.Sprintf("%.1f%%", ToFloat64(data["used_percent"])))
	printKeyValue("Active", FormatBytesValue(data["active"]))
	printKeyValue("Inactive", FormatBytesValue(data["inactive"]))
	printKeyValue("Buffers", FormatBytesValue(data["buffers"]))
	printKeyValue("Cached", FormatBytesValue(data["cached"]))
	fmt.Println()
}

func printSwapMemoryInfo(payload map[string]interface{}) {
	data := payload["swap_memory"].(map[string]interface{})

	fmt.Println(sysinfoSectionStyle.Render("=== Swap Memory ==="))
	printKeyValue("Total", FormatBytesValue(data["total"]))
	printKeyValue("Used", FormatBytesValue(data["used"]))
	printKeyValue("Free", FormatBytesValue(data["free"]))
	printKeyValue("Used Percent", fmt.Sprintf("%.1f%%", ToFloat64(data["used_percent"])))
	printKeyValue("Sin", FormatBytesValue(data["sin"]))
	printKeyValue("Sout", FormatBytesValue(data["sout"]))
	fmt.Println()
}

func printPartitionInfo(payload map[string]interface{}) {
	data := payload["partition"].(map[string]interface{})
	partitions, ok := data["partitions"].([]interface{})
	if !ok {
		return
	}

	fmt.Println(sysinfoSectionStyle.Render("=== Partitions ==="))
	fmt.Printf("  %-20s %-20s %-10s %10s %10s %10s %s\n",
		"Device", "Mountpoint", "Type", "Total", "Used", "Free", "Use%")
	fmt.Println()

	for _, pData := range partitions {
		if pMap, ok := pData.(map[string]interface{}); ok {
			device := ToString(pMap["device"])
			mountpoint := ToString(pMap["mountpoint"])
			fstype := ToString(pMap["fstype"])

			if errMsg, hasError := pMap["usage_error"]; hasError {
				fmt.Printf("  %-20s %-20s %-10s %s\n",
					device, mountpoint, fstype, sysinfoErrorStyle.Render(ToString(errMsg)))
			} else {
				total := FormatBytesCompact(pMap["total"])
				used := FormatBytesCompact(pMap["used"])
				free := FormatBytesCompact(pMap["free"])
				usedPercent := fmt.Sprintf("%.1f%%", ToFloat64(pMap["used_percent"]))

				fmt.Printf("  %-20s %-20s %-10s %10s %10s %10s %s\n",
					device, mountpoint, fstype, total, used, free, usedPercent)
			}
		}
	}
	fmt.Println()
}

func printIODiskInfo(payload map[string]interface{}) {
	data := payload["io_disk"].(map[string]interface{})

	fmt.Println(sysinfoSectionStyle.Render("=== Disk I/O ==="))
	for name, counterData := range data {
		if counterMap, ok := counterData.(map[string]interface{}); ok {
			fmt.Println(sysinfoTitleStyle.Render("  " + name + ":"))
			printKeyValue("Read Count", counterMap["read_count"])
			printKeyValue("Write Count", counterMap["write_count"])
			printKeyValue("Read Bytes", FormatBytesValue(counterMap["read_bytes"]))
			printKeyValue("Write Bytes", FormatBytesValue(counterMap["write_bytes"]))
			printKeyValue("Read Time", counterMap["read_time"])
			printKeyValue("Write Time", counterMap["write_time"])
		}
	}
	fmt.Println()
}

func printInterfacesInfo(payload map[string]interface{}) {
	data := payload["interfaces"].(map[string]interface{})
	interfaces, ok := data["interfaces"].([]interface{})
	if !ok {
		return
	}

	fmt.Println(sysinfoSectionStyle.Render("=== Network Interfaces ==="))
	for _, iData := range interfaces {
		if iMap, ok := iData.(map[string]interface{}); ok {
			fmt.Println(sysinfoTitleStyle.Render("  " + ToString(iMap["name"]) + ":"))
			printKeyValue("MTU", iMap["mtu"])
			printKeyValue("Flags", iMap["flags"])

			if addrs, ok := iMap["addresses"].([]interface{}); ok {
				addrStr := ""
				for j, addr := range addrs {
					if j > 0 {
						addrStr += ", "
					}
					addrStr += ToString(addr)
				}
				printKeyValue("Addresses", addrStr)
			}
		}
	}
	fmt.Println()
}

func printIONetworkInfo(payload map[string]interface{}) {
	data := payload["io_network"].(map[string]interface{})
	counters, ok := data["io_counters"].([]interface{})
	if !ok {
		return
	}

	fmt.Println(sysinfoSectionStyle.Render("=== Network I/O ==="))
	for _, cData := range counters {
		if cMap, ok := cData.(map[string]interface{}); ok {
			fmt.Println(sysinfoTitleStyle.Render("  " + ToString(cMap["name"]) + ":"))
			printKeyValue("Bytes Sent", FormatBytesValue(cMap["bytes_sent"]))
			printKeyValue("Bytes Recv", FormatBytesValue(cMap["bytes_recv"]))
			printKeyValue("Packets Sent", cMap["packets_sent"])
			printKeyValue("Packets Recv", cMap["packets_recv"])
			printKeyValue("Errors In", cMap["errin"])
			printKeyValue("Errors Out", cMap["errout"])
			printKeyValue("Dropped In", cMap["dropin"])
			printKeyValue("Dropped Out", cMap["dropout"])
		}
	}
	fmt.Println()
}

func printProcessesInfo(payload map[string]interface{}) {
	data := payload["processes"].(map[string]interface{})

	fmt.Println(sysinfoSectionStyle.Render("=== Processes ==="))
	printKeyValue("Total Count", data["count"])
	fmt.Println()

	procs, ok := data["processes"].([]interface{})
	if !ok {
		return
	}

	fmt.Printf("  %-8s %-30s %-10s %10s %10s %s\n",
		"PID", "Name", "Status", "CPU%", "Mem%", "Start Time")
	fmt.Println()

	for _, pData := range procs {
		if pMap, ok := pData.(map[string]interface{}); ok {
			pid := ToString(pMap["pid"])
			name := ToString(pMap["name"])
			status := ""
			if s, ok := pMap["status"].([]interface{}); ok && len(s) > 0 {
				status = ToString(s[0])
			}
			cpuPercent := fmt.Sprintf("%.1f", ToFloat64(pMap["cpu_percent"]))
			memPercent := fmt.Sprintf("%.1f", ToFloat64(pMap["memory_percent"]))
			createTime := ToString(pMap["create_time"])

			fmt.Printf("  %-8s %-30s %-10s %10s %10s %s\n",
				pid, name, status, cpuPercent, memPercent, createTime)
		}
	}
	fmt.Println()
}

func printKeyValue(key string, value interface{}) {
	fmt.Printf("  %s: %v\n", sysinfoKeyStyle.Render(key), sysinfoValueStyle.Render(fmt.Sprintf("%v", value)))
}

func FormatBytesValue(v interface{}) string {
	if v == nil {
		return "N/A"
	}
	return FormatBytesCompact(v)
}

func FormatBytesCompact(v interface{}) string {
	bytes := ToUint64(v)
	if bytes == 0 {
		return "0 B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	unitIndex := 0
	value := float64(bytes)

	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.1f %s", value, units[unitIndex])
}

func FormatUptimeValue(v interface{}) string {
	seconds := ToUint64(v)
	if seconds == 0 {
		return "0s"
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	result := ""
	if days > 0 {
		result += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm ", minutes)
	}
	result += fmt.Sprintf("%ds", secs)

	return result
}

func ToFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func ToUint64(v interface{}) uint64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return uint64(f)
	}
	if u, ok := v.(uint64); ok {
		return u
	}
	return 0
}

func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
