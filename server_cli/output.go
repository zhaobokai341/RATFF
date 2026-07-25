package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Base style definitions (globally initialized to avoid repeated memory allocation)
var (
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	styleDebug   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	stylePrompt  = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
)

// PrintSuccess outputs success messages in green.
func PrintSuccess(msg string) {
	prefix := styleSuccess.Render("[+]")
	fmt.Printf("%s %s\n", prefix, msg)
}

// PrintError outputs error messages in red.
func PrintError(msg string) {
	prefix := styleError.Render("[-]")
	fmt.Printf("%s %s\n", prefix, msg)
}

// PrintInfo outputs informational messages in blue.
func PrintInfo(msg string) {
	prefix := styleInfo.Render("[*]")
	fmt.Printf("%s %s\n", prefix, msg)
}

// PrintDebug outputs debug messages in gray.
func PrintDebug(msg string) {
	prefix := styleDebug.Render("[debug]")
	fmt.Printf("%s %s\n", prefix, msg)
}

// PrintWarn outputs warning messages in yellow.
func PrintWarn(msg string) {
	prefix := styleWarn.Render("[!]")
	fmt.Printf("%s %s\n", prefix, msg)
}

// BuildPrompt generates a colored CLI prompt string based on current mode.
func BuildPrompt(id string, inCommandMode bool) string {
	if id == "" {
		return stylePrompt.Render("(server) >> ")
	}
	if inCommandMode {
		return stylePrompt.Render(fmt.Sprintf("(%s)(command) >> ", id))
	}
	return stylePrompt.Render(fmt.Sprintf("(%s)(console) >> ", id))
}

// StyleCommandOutput renders command output in a bordered code block.
func StyleCommandOutput(output string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Render(output)
}
