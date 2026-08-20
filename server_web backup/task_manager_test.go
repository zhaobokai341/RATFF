package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskManagerCreateAndGet(t *testing.T) {
	tm := &TaskManager{
		tasks: make(map[string]*TransferTask),
	}

	task := tm.Create("task-1", "upload", 3, 1024)
	assert.NotNil(t, task)
	assert.Equal(t, "task-1", task.ID)
	assert.Equal(t, "upload", task.Type)
	assert.Equal(t, "pending", task.Status)
	assert.Equal(t, 3, task.FileCount)
	assert.Equal(t, int64(1024), task.TotalBytes)

	got := tm.Get("task-1")
	assert.NotNil(t, got)
	assert.Equal(t, task, got)

	missing := tm.Get("nonexistent")
	assert.Nil(t, missing)
}

func TestTaskManagerRemove(t *testing.T) {
	tm := &TaskManager{
		tasks: make(map[string]*TransferTask),
	}

	tm.Create("task-1", "upload", 1, 100)
	assert.NotNil(t, tm.Get("task-1"))

	tm.Remove("task-1")
	assert.Nil(t, tm.Get("task-1"))
}

func TestTransferTaskSubscribeAndUnsubscribe(t *testing.T) {
	task := &TransferTask{
		ID:        "task-1",
		Type:      "upload",
		Status:    "pending",
		listeners: make(map[chan TaskProgressEvent]struct{}),
	}

	ch := task.Subscribe()
	assert.NotNil(t, ch)

	task.mu.Lock()
	assert.Len(t, task.listeners, 1)
	task.mu.Unlock()

	task.Unsubscribe(ch)

	task.mu.Lock()
	assert.Len(t, task.listeners, 0)
	task.mu.Unlock()
}

func TestTransferTaskSetDone(t *testing.T) {
	task := &TransferTask{
		ID:        "task-1",
		Type:      "download",
		Status:    "pending",
		listeners: make(map[chan TaskProgressEvent]struct{}),
	}

	ch := task.Subscribe()

	go func() {
		task.SetDone("/tmp/result.zip")
	}()

	select {
	case evt := <-ch:
		assert.Equal(t, "done", evt.Type)
		assert.Equal(t, "/api/file/download_result?task_id=task-1", evt.DownloadURL)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for done event")
	}

	assert.Equal(t, "done", task.Status)
	assert.Equal(t, float64(100), task.Progress)
	assert.Equal(t, "/tmp/result.zip", task.ResultPath)
}

func TestTransferTaskSetError(t *testing.T) {
	task := &TransferTask{
		ID:        "task-1",
		Type:      "upload",
		Status:    "pending",
		listeners: make(map[chan TaskProgressEvent]struct{}),
	}

	ch := task.Subscribe()

	go func() {
		task.SetError(assert.AnError)
	}()

	select {
	case evt := <-ch:
		assert.Equal(t, "error", evt.Type)
		assert.Contains(t, evt.Message, assert.AnError.Error())
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for error event")
	}

	assert.Equal(t, "error", task.Status)
}

func TestTransferTaskUpdateProgress(t *testing.T) {
	task := &TransferTask{
		ID:         "task-1",
		Type:       "upload",
		Status:     "pending",
		TotalBytes: 1000,
		SentBytes:  0,
		FileCount:  3,
		FileName:   "test.txt",
		FileIndex:  1,
		listeners:  make(map[chan TaskProgressEvent]struct{}),
	}

	ch := task.Subscribe()

	task.SentBytes = 500
	go task.updateProgress()

	select {
	case evt := <-ch:
		assert.Equal(t, "progress", evt.Type)
		assert.Equal(t, float64(50), evt.Percent)
		assert.Equal(t, "test.txt", evt.FileName)
		assert.Equal(t, 1, evt.FileIndex)
		assert.Equal(t, 3, evt.FileCount)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for progress event")
	}

	assert.Equal(t, "processing", task.Status)
}

func TestTransferTaskProgressClamped(t *testing.T) {
	task := &TransferTask{
		ID:         "task-1",
		TotalBytes: 100,
		SentBytes:  200,
		listeners:  make(map[chan TaskProgressEvent]struct{}),
	}

	ch := task.Subscribe()
	go task.updateProgress()

	select {
	case evt := <-ch:
		assert.Equal(t, float64(100), evt.Percent)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for progress event")
	}
}

func TestTransferTaskBroadcastMultipleListeners(t *testing.T) {
	task := &TransferTask{
		ID:        "task-1",
		Status:    "pending",
		listeners: make(map[chan TaskProgressEvent]struct{}),
	}

	ch1 := task.Subscribe()
	ch2 := task.Subscribe()

	go task.SetDone("")

	select {
	case evt := <-ch1:
		assert.Equal(t, "done", evt.Type)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for listener 1")
	}

	select {
	case evt := <-ch2:
		assert.Equal(t, "done", evt.Type)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for listener 2")
	}
}
