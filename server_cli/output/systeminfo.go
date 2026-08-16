package output

import "fmt"

func PrintSystemInfoDetail(payload map[string]interface{}, fields []string, t func(string) string, tf func(string, ...interface{}) string) {
	clientID := ToString(payload["client_id"])
	PrintInfo(tf("systeminfo_client", clientID))

	if len(fields) == 0 {
		PrintInfo(t("systeminfo_all_fields"))
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
			printField(payload, field, printer, tf)
		} else {
			PrintError(tf("systeminfo_unknown_field", field))
		}
	}
}

func printField(payload map[string]interface{}, fieldName string, printer func(map[string]interface{}), tf func(string, ...interface{}) string) {
	if data, ok := payload[fieldName]; ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if errMsg, hasError := dataMap["error"]; hasError {
				printSectionError(fieldName, errMsg, tf)
				return
			}
		}
		printer(payload)
	}
}

func printSectionError(fieldName string, errMsg interface{}, tf func(string, ...interface{}) string) {
	fmt.Println(sysinfoSectionStyle.Render("=== " + fieldName + " ==="))
	fmt.Println(sysinfoErrorStyle.Render(tf("systeminfo_error_field", fieldName, errMsg)))
	fmt.Println()
}
