package main

import (
	"RATFF/server_cli/output"
)

// PrintSuccess prints a success message with styled prefix.
func PrintSuccess(msg string) {
	output.PrintSuccess(msg)
}

// PrintError prints an error message with styled prefix.
func PrintError(msg string) {
	output.PrintError(msg)
}

// PrintInfo prints an info message with styled prefix.
func PrintInfo(msg string) {
	output.PrintInfo(msg)
}

// PrintDebug prints a debug message with styled prefix.
func PrintDebug(msg string) {
	output.PrintDebug(msg)
}

// PrintWarn prints a warning message with styled prefix.
func PrintWarn(msg string) {
	output.PrintWarn(msg)
}

// BuildPrompt generates a styled CLI prompt.
func BuildPrompt(id string, inCommandMode bool) string {
	return output.BuildPrompt(id, inCommandMode)
}

// StyleCommandOutput wraps command output in a styled border.
func StyleCommandOutput(outputStr string) string {
	return output.StyleCommandOutput(outputStr)
}

// PrintCommandResult prints command execution results.
func PrintCommandResult(stdout, stderr string, exitCode int) {
	output.PrintCommandResult(stdout, stderr, exitCode, T, Tf)
}

// ProgressBar displays file transfer progress.
type ProgressBar = output.ProgressBar

// NewProgressBar creates a new progress bar instance.
func NewProgressBar(total int64, filename string) *ProgressBar {
	return output.NewProgressBar(total, filename)
}

// FormatBytes formats bytes into human-readable string.
func FormatBytes(b int64) string {
	return output.FormatBytes(b)
}

// formatBytes is an alias for backward compatibility.
func formatBytes(b int64) string {
	return output.FormatBytes(b)
}
