package main

import (
	"RATFF/shared"
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// readPassword reads a password from stdin, using terminal raw mode if available.
func readPassword() (string, error) {
	fd := int(os.Stdin.Fd())

	if !term.IsTerminal(fd) {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return scanner.Text(), nil
		}
		return "", scanner.Err()
	}

	password, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}

	fmt.Println()

	return string(password), nil
}

func main() {
	if err := shared.LoadLanguagePacks("lang"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load language packs: %v\n", err)
	}

	PrintInfo(T("prompt_path_password"))
	var err error
	cfg.PathPassword, err = readPassword()
	if err != nil {
		PrintError(Tf("input_read_error", err))
		os.Exit(1)
	}

	PrintInfo(T("prompt_login_password"))
	cfg.LoginPassword, err = readPassword()
	if err != nil {
		PrintError(Tf("input_read_error", err))
		os.Exit(1)
	}

	if cfg.LoginPassword == "" {
		PrintError(T("invalid_login_password"))
		os.Exit(1)
	}

	token, err := loginToAPI(cfg.LoginPassword)
	if err != nil {
		PrintError(Tf("login_failed", err))
		os.Exit(1)
	}
	jwtToken = token
	PrintSuccess(T("login_success"))

	wsConn, err := connectWS(getWSURL())
	if err != nil {
		PrintError(Tf("connect_failed", err))
		os.Exit(1)
	}
	defer wsConn.Close()

	PrintSuccess(T("connect_success"))

	go startResponseListener(getWSURL(), wsConn)

	inputScanner := bufio.NewScanner(os.Stdin)
	selectedID := ""
	inCommandMode := false

	for {
		prompt := buildPrompt(selectedID, inCommandMode)
		fmt.Print(prompt)

		if !inputScanner.Scan() {
			break
		}

		input := strings.TrimSpace(inputScanner.Text())
		if input == "" {
			continue
		}

		if selectedID == "" {
			selectedID = handleServerMode(input, selectedID)
		} else if inCommandMode {
			if input == "exit" {
				inCommandMode = false
				continue
			}
			handleCommandMode(input, selectedID)
		} else {
			action := handleConsoleMode(input, selectedID)
			switch action {
			case "enter_command":
				inCommandMode = true
			case "back":
				selectedID = ""
			case "exit":
				PrintSuccess(T("exited"))
				return
			}
		}
	}

	if err := inputScanner.Err(); err != nil {
		PrintError(Tf("input_read_error", err))
	}
}
