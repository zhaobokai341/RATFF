package output

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrintSuccess(t *testing.T) {
	PrintSuccess("test success message")
}

func TestPrintError(t *testing.T) {
	PrintError("test error message")
}

func TestPrintInfo(t *testing.T) {
	PrintInfo("test info message")
}

func TestPrintDebug(t *testing.T) {
	PrintDebug("test debug message")
}

func TestPrintWarn(t *testing.T) {
	PrintWarn("test warning message")
}

func TestBuildPromptNoID(t *testing.T) {
	prompt := BuildPrompt("", false)
	assert.Contains(t, prompt, "server")
}

func TestBuildPromptWithIDCommandMode(t *testing.T) {
	prompt := BuildPrompt("client-1", true)
	assert.Contains(t, prompt, "client-1")
	assert.Contains(t, prompt, "command")
}

func TestBuildPromptWithIDConsoleMode(t *testing.T) {
	prompt := BuildPrompt("client-1", false)
	assert.Contains(t, prompt, "client-1")
	assert.Contains(t, prompt, "console")
}

func TestStyleCommandOutput(t *testing.T) {
	output := StyleCommandOutput("test output")
	assert.NotEmpty(t, output)
}

func TestPrintCommandResultSuccess(t *testing.T) {
	tFunc := func(key string) string { return key }
	tfFunc := func(key string, args ...interface{}) string { return key }

	PrintCommandResult("hello world", "", 0, tFunc, tfFunc)
}

func TestPrintCommandResultWithError(t *testing.T) {
	tFunc := func(key string) string { return key }
	tfFunc := func(key string, args ...interface{}) string { return key }

	PrintCommandResult("", "error occurred", 1, tFunc, tfFunc)
}

func TestPrintCommandResultWithBoth(t *testing.T) {
	tFunc := func(key string) string { return key }
	tfFunc := func(key string, args ...interface{}) string { return key }

	PrintCommandResult("stdout output", "stderr output", 0, tFunc, tfFunc)
}

func TestProgressBarNew(t *testing.T) {
	bar := NewProgressBar(1000, "test.txt")
	assert.Equal(t, int64(1000), bar.total)
	assert.Equal(t, "test.txt", bar.filename)
	assert.False(t, bar.done)
}

func TestProgressBarAdd(t *testing.T) {
	bar := NewProgressBar(1000, "test.txt")
	bar.Add(100)
	bar.mu.Lock()
	assert.Equal(t, int64(100), bar.current)
	bar.mu.Unlock()
}

func TestProgressBarSetTotal(t *testing.T) {
	bar := NewProgressBar(1000, "test.txt")
	bar.SetTotal(2000)
	bar.mu.Lock()
	assert.Equal(t, int64(2000), bar.total)
	bar.mu.Unlock()
}

func TestProgressBarMarkDone(t *testing.T) {
	bar := NewProgressBar(1000, "test.txt")
	bar.MarkDone()
	bar.mu.Lock()
	assert.True(t, bar.done)
	bar.mu.Unlock()
}

func TestProgressBarDisplay(t *testing.T) {
	bar := NewProgressBar(1000, "test.txt")
	bar.Add(500)
	bar.Display()
}

func TestProgressBarDisplayDone(t *testing.T) {
	bar := NewProgressBar(1000, "test.txt")
	bar.Add(1000)
	bar.MarkDone()
	bar.Display()
}

func TestProgressBarDisplayZeroTotal(t *testing.T) {
	bar := NewProgressBar(0, "test.txt")
	bar.Display()
}

func TestProgressBarDisplayOver100Percent(t *testing.T) {
	bar := NewProgressBar(100, "test.txt")
	bar.Add(200)
	bar.Display()
}

func TestFormatBytes(t *testing.T) {
	assert.Equal(t, "0 B", FormatBytes(0))
	assert.Equal(t, "100 B", FormatBytes(100))
	assert.Equal(t, "1.0 KB", FormatBytes(1024))
	assert.Equal(t, "1.5 KB", FormatBytes(1536))
	assert.Equal(t, "1.0 MB", FormatBytes(1024*1024))
	assert.Equal(t, "1.0 GB", FormatBytes(1024*1024*1024))
}

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "0s", FormatDuration(0))
	assert.Equal(t, "30s", FormatDuration(30*time.Second))
	assert.Equal(t, "1m30s", FormatDuration(90*time.Second))
	assert.Equal(t, "1h30m", FormatDuration(90*time.Minute))
}
