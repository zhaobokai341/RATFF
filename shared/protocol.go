package shared

import "time"

// MessageType defines the type of message in the protocol.
type MessageType string

const (
	MsgRegister  MessageType = "register"
	MsgHeartbeat MessageType = "heartbeat"
	MsgCommand   MessageType = "command"
	MsgResponse  MessageType = "response"
	MsgError     MessageType = "error"
)

// CommandType defines the type of command to execute.
type CommandType string

const (
	CmdScreenCapture        CommandType = "screen_capture"
	CmdShellExec            CommandType = "shell_exec"
	CmdShellExecBg          CommandType = "shell_exec_bg"
	CmdFileList             CommandType = "file_list"
	CmdFileUpload           CommandType = "file_upload"
	CmdFileUploadStart      CommandType = "file_upload_start"
	CmdFileUploadChunk      CommandType = "file_upload_chunk"
	CmdFileUploadComplete   CommandType = "file_upload_complete"
	CmdFileDownload         CommandType = "file_download"
	CmdFileDownloadStart    CommandType = "file_download_start"
	CmdFileDownloadChunk    CommandType = "file_download_chunk"
	CmdFileDownloadComplete CommandType = "file_download_complete"
	CmdFileMove             CommandType = "file_move"
	CmdFileDelete           CommandType = "file_delete"
	CmdSystemInfo           CommandType = "system_info"
	CmdExit                 CommandType = "exit"
	CmdCd                   CommandType = "cd"
)

// Message represents a protocol message exchanged between components.
type Message struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	Command   CommandType            `json:"command,omitempty"`
	ClientID  string                 `json:"client_id,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// NewMessage creates a new protocol message with generated ID and timestamp.
func NewMessage(msgType MessageType, cmd CommandType, clientID string, payload map[string]interface{}) Message {
	return Message{
		ID:        GenerateID(),
		Type:      msgType,
		Command:   cmd,
		ClientID:  clientID,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
	}
}
