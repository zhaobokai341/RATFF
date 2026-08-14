package main

import (
	"RATFF/server_cli/output"
)

// printSystemInfoDetail displays system information in a formatted layout.
func printSystemInfoDetail(payload map[string]interface{}, fields []string) {
	output.PrintSystemInfoDetail(payload, fields, T, Tf)
}
