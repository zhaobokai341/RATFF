package main

import (
	"RATFF/server_cli/output"
)

// HelpCommand represents a help command entry with name and description.
type HelpCommand = output.HelpCommand

// PrintHelp outputs a formatted list of available commands.
func PrintHelp(commands []HelpCommand) {
	output.PrintHelp(commands, T)
}
