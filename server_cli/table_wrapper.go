package main

import (
	"RATFF/server_cli/output"
	"RATFF/shared"
)

// PrintClientTable displays connected clients in a formatted table.
func PrintClientTable(clients []shared.ClientInfo) {
	output.PrintClientTable(clients, T, Tf)
}

// PrintFileTable displays file list in a formatted table.
func PrintFileTable(currentPath string, files []interface{}) {
	output.PrintFileTable(currentPath, files, T, Tf)
}

// formatID formats client ID to fit column width.
func formatID(id string) string {
	return output.FormatID(id)
}
