package main

import (
	"fmt"
	"os"
	"strings"
)

// buildPrompt generates the CLI prompt based on current mode.
func buildPrompt(id string, inCommandMode bool) string {
	return BuildPrompt(id, inCommandMode)
}

// handleServerMode processes commands in server mode.
func handleServerMode(input string, selectedID string) string {
	parts := strings.Fields(input)
	cmd := parts[0]

	switch cmd {
	case "help":
		printServerHelp()
	case "list":
		listClients()
	case "select":
		if len(parts) < 2 {
			PrintError(T("usage_select"))
			return selectedID
		}
		if selectClient(parts[1]) {
			return parts[1]
		}
	case "delete":
		if len(parts) < 2 {
			PrintError(T("usage_delete"))
			return selectedID
		}
		deleteClient(parts[1])
	case "clear":
		clearScreen()
	case "exit":
		PrintSuccess(T("exited"))
		os.Exit(0)
	default:
		PrintError(T("invalid_command"))
	}

	return selectedID
}

// handleConsoleMode processes commands in console mode.
func handleConsoleMode(input string) string {
	parts := strings.Fields(input)
	cmd := parts[0]

	switch cmd {
	case "help":
		printConsoleHelp()
		return ""
	case "command":
		return "enter_command"
	case "back":
		return "back"
	case "exit":
		return "exit"
	default:
		PrintError(T("invalid_command"))
		return ""
	}
}

// printServerHelp displays available commands in server mode.
func printServerHelp() {
	PrintHelp([]HelpCommand{
		{"list", T("cmd_list_desc")},
		{"select <id>", T("cmd_select_desc")},
		{"delete <id>", T("cmd_delete_desc")},
		{"clear", T("cmd_clear_desc")},
		{"help", T("cmd_help_desc")},
		{"exit", T("cmd_exit_desc")},
	})
}

// printConsoleHelp displays available commands in console mode.
func printConsoleHelp() {
	PrintHelp([]HelpCommand{
		{"command", T("cmd_command_desc")},
		{"back", T("cmd_back_desc")},
		{"help", T("cmd_help_desc")},
		{"exit", T("cmd_exit_desc")},
	})
}

// clearScreen clears the terminal screen.
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// handleCommandMode executes a shell command on the selected client.
func handleCommandMode(input string, id string) {
	sendShellCommand(id, input)
}
