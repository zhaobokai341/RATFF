package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	sysinfoTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#8B5CF6")).
				MarginBottom(1)

	sysinfoSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#60A5FA")).
				MarginTop(1).
				MarginBottom(1)

	sysinfoKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			Width(18).
			Align(lipgloss.Right)

	sysinfoValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5E7EB"))

	sysinfoErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444"))
)

func printSystemInfoDetail(payload map[string]interface{}, fields []string) {
	clientID := toString(payload["client_id"])
	PrintInfo(Tf("systeminfo_client", clientID))

	if len(fields) == 0 {
		PrintInfo(T("systeminfo_all_fields"))
		fields = []string{"host", "cpu", "memory", "swap_memory", "partition", "io_disk", "interfaces", "io_network", "processes"}
	}

	fieldPrinters := map[string]func(map[string]interface{}){
		"host":        printHostInfo,
		"cpu":         printCPUInfo,
		"memory":      printMemoryInfo,
		"swap_memory": printSwapMemoryInfo,
		"partition":   printPartitionInfo,
		"io_disk":     printIODiskInfo,
		"interfaces":  printInterfacesInfo,
		"io_network":  printIONetworkInfo,
		"processes":   printProcessesInfo,
	}

	for _, field := range fields {
		if printer, ok := fieldPrinters[field]; ok {
			printField(payload, field, printer)
		} else {
			PrintError(Tf("systeminfo_unknown_field", field))
		}
	}
}

func printField(payload map[string]interface{}, fieldName string, printer func(map[string]interface{})) {
	if data, ok := payload[fieldName]; ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if errMsg, hasError := dataMap["error"]; hasError {
				printSectionError(fieldName, errMsg)
				return
			}
		}
		printer(payload)
	}
}

func printSectionError(fieldName string, errMsg interface{}) {
	fmt.Println(sysinfoSectionStyle.Render("=== " + fieldName + " ==="))
	fmt.Println(sysinfoErrorStyle.Render(Tf("systeminfo_error_field", fieldName, errMsg)))
	fmt.Println()
}

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
	printKeyValue("Uptime", formatUptimeValue(data["uptime"]))
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
	printKeyValue("Total", formatBytesValue(data["total"]))
	printKeyValue("Used", formatBytesValue(data["used"]))
	printKeyValue("Available", formatBytesValue(data["available"]))
	printKeyValue("Free", formatBytesValue(data["free"]))
	printKeyValue("Used Percent", fmt.Sprintf("%.1f%%", toFloat64(data["used_percent"])))
	printKeyValue("Active", formatBytesValue(data["active"]))
	printKeyValue("Inactive", formatBytesValue(data["inactive"]))
	printKeyValue("Buffers", formatBytesValue(data["buffers"]))
	printKeyValue("Cached", formatBytesValue(data["cached"]))
	fmt.Println()
}

func printSwapMemoryInfo(payload map[string]interface{}) {
	data := payload["swap_memory"].(map[string]interface{})

	fmt.Println(sysinfoSectionStyle.Render("=== Swap Memory ==="))
	printKeyValue("Total", formatBytesValue(data["total"]))
	printKeyValue("Used", formatBytesValue(data["used"]))
	printKeyValue("Free", formatBytesValue(data["free"]))
	printKeyValue("Used Percent", fmt.Sprintf("%.1f%%", toFloat64(data["used_percent"])))
	printKeyValue("Sin", formatBytesValue(data["sin"]))
	printKeyValue("Sout", formatBytesValue(data["sout"]))
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
			device := toString(pMap["device"])
			mountpoint := toString(pMap["mountpoint"])
			fstype := toString(pMap["fstype"])

			if errMsg, hasError := pMap["usage_error"]; hasError {
				fmt.Printf("  %-20s %-20s %-10s %s\n",
					device, mountpoint, fstype, sysinfoErrorStyle.Render(toString(errMsg)))
			} else {
				total := formatBytesCompact(pMap["total"])
				used := formatBytesCompact(pMap["used"])
				free := formatBytesCompact(pMap["free"])
				usedPercent := fmt.Sprintf("%.1f%%", toFloat64(pMap["used_percent"]))

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
			printKeyValue("Read Bytes", formatBytesValue(counterMap["read_bytes"]))
			printKeyValue("Write Bytes", formatBytesValue(counterMap["write_bytes"]))
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
			fmt.Println(sysinfoTitleStyle.Render("  " + toString(iMap["name"]) + ":"))
			printKeyValue("MTU", iMap["mtu"])
			printKeyValue("Flags", iMap["flags"])

			if addrs, ok := iMap["addresses"].([]interface{}); ok {
				addrStr := ""
				for j, addr := range addrs {
					if j > 0 {
						addrStr += ", "
					}
					addrStr += toString(addr)
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
			fmt.Println(sysinfoTitleStyle.Render("  " + toString(cMap["name"]) + ":"))
			printKeyValue("Bytes Sent", formatBytesValue(cMap["bytes_sent"]))
			printKeyValue("Bytes Recv", formatBytesValue(cMap["bytes_recv"]))
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
			pid := toString(pMap["pid"])
			name := toString(pMap["name"])
			status := ""
			if s, ok := pMap["status"].([]interface{}); ok && len(s) > 0 {
				status = toString(s[0])
			}
			cpuPercent := fmt.Sprintf("%.1f", toFloat64(pMap["cpu_percent"]))
			memPercent := fmt.Sprintf("%.1f", toFloat64(pMap["memory_percent"]))
			createTime := toString(pMap["create_time"])

			fmt.Printf("  %-8s %-30s %-10s %10s %10s %s\n",
				pid, name, status, cpuPercent, memPercent, createTime)
		}
	}
	fmt.Println()
}

func printKeyValue(key string, value interface{}) {
	fmt.Printf("  %s: %v\n", sysinfoKeyStyle.Render(key), sysinfoValueStyle.Render(fmt.Sprintf("%v", value)))
}

func formatBytesValue(v interface{}) string {
	if v == nil {
		return "N/A"
	}
	return formatBytesCompact(v)
}

func formatBytesCompact(v interface{}) string {
	bytes := toUint64(v)
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

func formatUptimeValue(v interface{}) string {
	seconds := toUint64(v)
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

func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func toUint64(v interface{}) uint64 {
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

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
