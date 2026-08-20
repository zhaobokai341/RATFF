package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type TransferTask struct {
	ID         string
	Type       string
	Status     string
	Progress   float64
	FileName   string
	FileIndex  int
	FileCount  int
	TotalBytes int64
	SentBytes  int64
	Error      string
	ResultPath string

	mu        sync.Mutex
	listeners map[chan TaskProgressEvent]struct{}
}

type TaskProgressEvent struct {
	Type        string  `json:"type"`
	Stage       string  `json:"stage,omitempty"`
	FileName    string  `json:"file_name,omitempty"`
	FileIndex   int     `json:"file_index,omitempty"`
	FileCount   int     `json:"file_count,omitempty"`
	Percent     float64 `json:"percent"`
	Message     string  `json:"message,omitempty"`
	DownloadURL string  `json:"download_url,omitempty"`
}

type TaskManager struct {
	mu    sync.RWMutex
	tasks map[string]*TransferTask
}

var taskManager = &TaskManager{
	tasks: make(map[string]*TransferTask),
}

func (tm *TaskManager) Create(id, taskType string, fileCount int, totalBytes int64) *TransferTask {
	task := &TransferTask{
		ID:         id,
		Type:       taskType,
		Status:     "pending",
		FileCount:  fileCount,
		TotalBytes: totalBytes,
		listeners:  make(map[chan TaskProgressEvent]struct{}),
	}
	tm.mu.Lock()
	tm.tasks[id] = task
	tm.mu.Unlock()
	return task
}

func (tm *TaskManager) Get(id string) *TransferTask {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.tasks[id]
}

func (tm *TaskManager) Remove(id string) {
	tm.mu.Lock()
	delete(tm.tasks, id)
	tm.mu.Unlock()
}

func (t *TransferTask) updateProgress() {
	t.mu.Lock()
	t.Status = "processing"
	if t.TotalBytes > 0 {
		t.Progress = float64(t.SentBytes) / float64(t.TotalBytes) * 100
		if t.Progress > 100 {
			t.Progress = 100
		}
	}
	t.mu.Unlock()

	t.broadcast(TaskProgressEvent{
		Type:      "progress",
		Stage:     t.Type + "ing",
		FileName:  t.FileName,
		FileIndex: t.FileIndex,
		FileCount: t.FileCount,
		Percent:   t.Progress,
	})
}

func (t *TransferTask) SetDone(resultPath string) {
	t.mu.Lock()
	t.Status = "done"
	t.Progress = 100
	t.ResultPath = resultPath
	t.mu.Unlock()

	t.broadcast(TaskProgressEvent{
		Type:        "done",
		DownloadURL: "/api/file/download_result?task_id=" + t.ID,
	})
}

func (t *TransferTask) SetError(err error) {
	t.mu.Lock()
	t.Status = "error"
	t.Error = err.Error()
	t.mu.Unlock()

	t.broadcast(TaskProgressEvent{
		Type:    "error",
		Message: err.Error(),
	})
}

func (t *TransferTask) Subscribe() chan TaskProgressEvent {
	ch := make(chan TaskProgressEvent, 10)
	t.mu.Lock()
	t.listeners[ch] = struct{}{}
	t.mu.Unlock()
	return ch
}

func (t *TransferTask) Unsubscribe(ch chan TaskProgressEvent) {
	t.mu.Lock()
	delete(t.listeners, ch)
	t.mu.Unlock()
}

func (t *TransferTask) broadcast(evt TaskProgressEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for ch := range t.listeners {
		select {
		case ch <- evt:
		default:
		}
	}
}

func handleTaskProgress(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(400, gin.H{"error": "missing task_id"})
		return
	}

	task := taskManager.Get(taskID)
	if task == nil {
		c.JSON(404, gin.H{"error": "task not found"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Check if task is already completed before subscribing
	task.mu.Lock()
	status := task.Status
	taskErr := task.Error
	task.mu.Unlock()

	if status == "done" {
		evt := TaskProgressEvent{
			Type:        "done",
			DownloadURL: "/api/file/download_result?task_id=" + taskID,
		}
		data, _ := json.Marshal(evt)
		fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", string(data))
		c.Writer.Flush()
		return
	}

	if status == "error" {
		evt := TaskProgressEvent{
			Type:    "error",
			Message: taskErr,
		}
		data, _ := json.Marshal(evt)
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", string(data))
		c.Writer.Flush()
		return
	}

	ch := task.Subscribe()
	defer task.Unsubscribe(ch)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case evt := <-ch:
			data, _ := json.Marshal(evt)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, string(data))
			return evt.Type != "done" && evt.Type != "error"
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

func handleTaskStatus(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(400, gin.H{"error": "missing task_id"})
		return
	}

	task := taskManager.Get(taskID)
	if task == nil {
		c.JSON(404, gin.H{"error": "task not found"})
		return
	}

	task.mu.Lock()
	status := task.Status
	progress := task.Progress
	fileName := task.FileName
	fileIndex := task.FileIndex
	fileCount := task.FileCount
	taskErr := task.Error
	task.mu.Unlock()

	resp := gin.H{
		"status":     status,
		"percent":    progress,
		"file_name":  fileName,
		"file_index": fileIndex,
		"file_count": fileCount,
	}

	if status == "error" {
		resp["error"] = taskErr
	}

	c.JSON(200, resp)
}

func handleDownloadResult(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(400, gin.H{"error": "missing task_id"})
		return
	}

	task := taskManager.Get(taskID)
	if task == nil {
		c.JSON(404, gin.H{"error": "task not found"})
		return
	}

	if task.Status != "done" {
		c.JSON(400, gin.H{"error": "task not completed"})
		return
	}

	if task.ResultPath == "" {
		c.JSON(404, gin.H{"error": "no result file"})
		return
	}

	file, err := os.Open(task.ResultPath)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to open result file"})
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to stat result file"})
		return
	}

	resultPath := task.ResultPath
	defer func() {
		info, err := os.Stat(resultPath)
		if err == nil && info.IsDir() {
			if rmErr := os.RemoveAll(resultPath); rmErr != nil {
				log.WithError(rmErr).Warn("Failed to remove result directory")
			}
		} else {
			if rmErr := os.Remove(resultPath); rmErr != nil {
				log.WithError(rmErr).Warn("Failed to remove result file")
			}
		}
		tmpDir := filepath.Dir(resultPath)
		if tmpDir != "" {
			entries, err := os.ReadDir(tmpDir)
			if err == nil && len(entries) == 0 {
				if rmErr := os.Remove(tmpDir); rmErr != nil {
					log.WithError(rmErr).Warn("Failed to remove empty temp directory")
				}
			}
		}
	}()
	taskManager.Remove(taskID)

	safeName := filepath.Base(task.FileName)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", safeName, safeName))
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	http.ServeContent(c.Writer, c.Request, safeName, fileInfo.ModTime(), file)
}
