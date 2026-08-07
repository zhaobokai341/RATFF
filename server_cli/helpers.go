package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/shlex"
)

// buildPrompt generates the CLI prompt based on current mode.
func buildPrompt(id string, inCommandMode bool) string {
	return BuildPrompt(id, inCommandMode)
}

// handleServerMode processes commands in server mode (list, select, delete, etc.).
func handleServerMode(input string, selectedID string) string {
	args, err := shlex.Split(input)
	if err != nil {
		PrintError(Tf("invalid_command_args", err))
		return selectedID
	}

	if len(args) == 0 {
		return selectedID
	}

	cmd := args[0]

	switch cmd {
	case "help":
		printServerHelp()
	case "list":
		listClients()
	case "select":
		if len(args) < 2 {
			PrintError(T("usage_select"))
			return selectedID
		}
		if selectClient(args[1]) {
			return args[1]
		}
	case "delete":
		if len(args) < 2 {
			PrintError(T("usage_delete"))
			return selectedID
		}
		deleteClient(args[1])
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

// handleConsoleMode processes commands in console mode (command, cd, bg, back, exit).
func handleConsoleMode(input string, selectedID string) string {
	args, err := shlex.Split(input)
	if err != nil {
		PrintError(Tf("invalid_command_args", err))
		return ""
	}

	if len(args) == 0 {
		return ""
	}

	cmd := args[0]

	switch cmd {
	case "help":
		printConsoleHelp()
		return ""
	case "command":
		return "enter_command"
	case "cd":
		if len(args) < 2 {
			PrintError(T("usage_cd"))
			return ""
		}
		dir := strings.Join(args[1:], " ")
		cdClient(selectedID, dir)
		return ""
	case "upload":
		if len(args) < 2 {
			PrintError(T("usage_upload"))
			return ""
		}
		localPath := args[1]
		var remotePath string
		if len(args) >= 3 {
			remotePath = args[2]
		}
		uploadFile(selectedID, localPath, remotePath)
		return ""
	case "download":
		if len(args) < 2 {
			PrintError(T("usage_download"))
			return ""
		}
		remotePath := args[1]
		var localPath string
		if len(args) >= 3 {
			localPath = args[2]
		}
		downloadFile(selectedID, remotePath, localPath)
		return ""
	case "back":
		return "back"
	case "exit":
		return "exit"
	case "bg":
		return handleConsoleBgCommand(args, selectedID)
	case "systeminfo":
		var fields []string
		if len(args) > 1 {
			fields = args[1:]
		}
		systeminfo(selectedID, fields)
		return ""
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
		{"cd <dir>", T("cmd_cd_console_desc")},
		{"upload <local> [remote]", T("cmd_upload_desc")},
		{"download <remote> [local]", T("cmd_download_desc")},
		{"bg <cmd> [file]", T("cmd_bg_desc")},
		{"systeminfo [fields...]", T("cmd_systeminfo_desc")},
		{"back", T("cmd_back_desc")},
		{"help", T("cmd_help_desc")},
		{"exit", T("cmd_exit_desc")},
	})
}

// clearScreen clears the terminal screen using ANSI escape codes.
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// handleCommandMode executes a shell command on the selected client.
func handleCommandMode(input string, id string) {
	sendShellCommand(id, input)
}

// handleConsoleBgCommand processes bg command in console mode.
func handleConsoleBgCommand(args []string, selectedID string) string {
	if len(args) < 2 {
		PrintError(T("usage_bg"))
		return ""
	}

	bgArgs := args[1:]
	var outputFile string
	var cmdParts []string

	for i, arg := range bgArgs {
		if strings.HasPrefix(arg, "/") || strings.Contains(arg, ":\\") {
			outputFile = arg
			cmdParts = bgArgs[:i]
			break
		}
	}

	if outputFile == "" {
		cmdParts = bgArgs
	}

	cmd := strings.Join(cmdParts, " ")
	if cmd == "" {
		PrintError(T("usage_bg"))
		return ""
	}

	id := selectedID
	if id == "" {
		clients, err := fetchClients()
		if err != nil {
			PrintError(Tf("fetch_clients_failed", err))
			return ""
		}

		if len(clients) == 0 {
			PrintError(T("no_clients"))
			return ""
		}

		id = clients[0].ID
	}

	sendBgCommand(id, cmd, outputFile)
	return ""
}
