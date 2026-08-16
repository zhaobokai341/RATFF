package main

import (
	"os"
	"os/exec"
	"runtime"
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
	case "clear":
		clearScreen()
		return ""
	case "cd":
		if len(args) < 2 {
			PrintError(T("usage_cd"))
			return ""
		}
		dir := strings.Join(args[1:], " ")
		cdClient(selectedID, dir)
		return ""
	case "pwd":
		pwdClient(selectedID)
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
	case "ls":
		var path string
		if len(args) >= 2 {
			path = strings.Join(args[1:], " ")
		}
		listFiles(selectedID, path)
		return ""
	case "mv":
		if len(args) < 3 {
			PrintError(T("usage_mv"))
			return ""
		}
		originPath := args[1]
		newPath := strings.Join(args[2:], " ")
		moveFile(selectedID, originPath, newPath)
		return ""
	case "cp":
		if len(args) < 3 {
			PrintError(T("usage_cp"))
			return ""
		}
		originPath := args[1]
		newPath := strings.Join(args[2:], " ")
		copyRemoteFile(selectedID, originPath, newPath)
		return ""
	case "rm":
		if len(args) < 2 {
			PrintError(T("usage_rm"))
			return ""
		}
		path := strings.Join(args[1:], " ")
		deleteFile(selectedID, path)
		return ""
	default:
		PrintError(T("invalid_command"))
		return ""
	}
}

// printServerHelp displays available commands in server mode.
func printServerHelp() {
	PrintHelp([]HelpCommand{
		{Cmd: "list", Desc: T("cmd_list_desc")},
		{Cmd: "select <id>", Desc: T("cmd_select_desc")},
		{Cmd: "delete <id>", Desc: T("cmd_delete_desc")},
		{Cmd: "clear", Desc: T("cmd_clear_desc")},
		{Cmd: "help", Desc: T("cmd_help_desc")},
		{Cmd: "exit", Desc: T("cmd_exit_desc")},
	})
}

// printConsoleHelp displays available commands in console mode.
func printConsoleHelp() {
	PrintHelp([]HelpCommand{
		{Cmd: "command", Desc: T("cmd_command_desc")},
		{Cmd: "cd <dir>", Desc: T("cmd_cd_console_desc")},
		{Cmd: "pwd", Desc: T("cmd_pwd_desc")},
		{Cmd: "ls [path]", Desc: T("cmd_ls_desc")},
		{Cmd: "mv <origin> <destination>", Desc: T("cmd_mv_desc")},
		{Cmd: "cp <origin> <destination>", Desc: T("cmd_cp_desc")},
		{Cmd: "rm <path>", Desc: T("cmd_rm_desc")},
		{Cmd: "upload <local> [remote]", Desc: T("cmd_upload_desc")},
		{Cmd: "download <remote> [local]", Desc: T("cmd_download_desc")},
		{Cmd: "bg <cmd> [file]", Desc: T("cmd_bg_desc")},
		{Cmd: "systeminfo [fields...]", Desc: T("cmd_systeminfo_desc")},
		{Cmd: "back", Desc: T("cmd_back_desc")},
		{Cmd: "help", Desc: T("cmd_help_desc")},
		{Cmd: "exit", Desc: T("cmd_exit_desc")},
	})
}

// clearScreen clears the terminal screen using ANSI escape codes.
func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
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
