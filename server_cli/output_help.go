package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Help command style definitions (globally initialized)
var (
	helpCmdStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Width(18)
	helpDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// HelpCommand represents a help command entry with name and description.
type HelpCommand struct {
	Cmd  string
	Desc string
}

// PrintHelp outputs a formatted list of available commands.
func PrintHelp(commands []HelpCommand) {
	PrintInfo(T("available_commands"))
	fmt.Println()

	for _, cmd := range commands {
		cmdStr := helpCmdStyle.Render(cmd.Cmd)
		descStr := helpDescStyle.Render("- " + cmd.Desc)
		fmt.Printf("  %s %s\n", cmdStr, descStr)
	}
}
