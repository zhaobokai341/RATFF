package main

import (
	"RATFF/server_cli/api"
	"RATFF/server_cli/client"
	"RATFF/server_cli/output"

	"github.com/gorilla/websocket"
)

// jwtToken stores the authenticated JWT token for API requests.
var jwtToken string

// apiClient is the global API client instance.
var apiClient *api.Client

// wsManager is the global WebSocket manager.
var wsManager *api.WebSocketManager

// wsConn is the current WebSocket connection.
var wsConn *websocket.Conn

// clientManager is the global client operations manager.
var clientManager *client.Manager

// initPrintFuncs initializes the print functions for the client manager.
func initPrintFuncs() client.PrintFuncs {
	return client.PrintFuncs{
		Success:     output.PrintSuccess,
		Error:       output.PrintError,
		Info:        output.PrintInfo,
		Warn:        output.PrintWarn,
		FormatBytes: output.FormatBytes,
	}
}
