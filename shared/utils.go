package shared

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AsyncWriter wraps an io.Writer with asynchronous writes via a buffered channel.
type AsyncWriter struct {
	ch   chan []byte
	done chan struct{}
	wg   sync.WaitGroup
	out  io.Writer
	mu   sync.Mutex
}

// NewAsyncWriter creates a new AsyncWriter with the specified buffer size.
func NewAsyncWriter(out io.Writer, bufferSize int) *AsyncWriter {
	aw := &AsyncWriter{
		ch:   make(chan []byte, bufferSize),
		done: make(chan struct{}),
		out:  out,
	}
	aw.wg.Add(1)
	go aw.run()
	return aw
}

// run is the background goroutine that writes log data to the underlying writer.
func (aw *AsyncWriter) run() {
	defer aw.wg.Done()
	for {
		select {
		case data := <-aw.ch:
			aw.mu.Lock()
			if _, err := aw.out.Write(data); err != nil {
				fmt.Fprintf(os.Stderr, "AsyncWriter write error: %v\n", err)
			}
			aw.mu.Unlock()
		case <-aw.done:
			// Flush remaining messages
			for len(aw.ch) > 0 {
				data := <-aw.ch
				aw.mu.Lock()
				if _, err := aw.out.Write(data); err != nil {
					fmt.Fprintf(os.Stderr, "AsyncWriter flush error: %v\n", err)
				}
				aw.mu.Unlock()
			}
			return
		}
	}
}

// Write implements io.Writer interface. It sends data to the channel asynchronously.
func (aw *AsyncWriter) Write(p []byte) (n int, err error) {
	// Make a copy of the data since the caller may reuse the buffer
	data := make([]byte, len(p))
	copy(data, p)

	select {
	case aw.ch <- data:
		return len(p), nil
	default:
		// Channel is full, write directly with lock to avoid blocking
		aw.mu.Lock()
		defer aw.mu.Unlock()
		return aw.out.Write(p)
	}
}

// Close gracefully shuts down the async writer, flushing remaining messages.
func (aw *AsyncWriter) Close() {
	close(aw.done)
	aw.wg.Wait()
}

// InitLogger creates a configured logrus.Entry with the specified level and format.
func InitLogger(level string, format string) *logrus.Entry {
	logger, _ := InitLoggerWithWriter(level, format, false)
	return logger
}

// InitLoggerAsync creates a configured logrus.Entry with optional async output.
// When async is true, log writes are buffered and processed by a background goroutine.
func InitLoggerAsync(level string, format string, async bool) *logrus.Entry {
	logger, _ := InitLoggerWithWriter(level, format, async)
	return logger
}

// InitLoggerWithWriter creates a configured logrus.Entry with optional async output.
// Returns the logger entry and the async writer (if async is enabled).
func InitLoggerWithWriter(level string, format string, async bool) (*logrus.Entry, *AsyncWriter) {
	logger := logrus.New()

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	if format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	var asyncWriter *AsyncWriter
	if async {
		// Use async writer with 10000 message buffer
		asyncWriter = NewAsyncWriter(os.Stderr, 10000)
		logger.SetOutput(asyncWriter)
		logger.SetNoLock() // No need for locks with async output
	}

	return logger.WithField("service", "ratff"), asyncWriter
}

// GenerateID returns a UUID string.
func GenerateID() string {
	return uuid.New().String()
}

// EncryptAES encrypts plaintext using AES-GCM with a random nonce.
// Key must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256.
func EncryptAES(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key size: %d bytes, must be 16, 24, or 32 bytes", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptAES decrypts ciphertext using AES-GCM.
// Key must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256.
func DecryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key size: %d bytes, must be 16, 24, or 32 bytes", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GetEnv returns the value of an environment variable, or a fallback if not set.
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// IsDebugEnv returns true if running in debug/development environment.
func IsDebugEnv() bool {
	env := GetEnv("APP_ENV", "debug")
	return env == "debug" || env == "development" || env == "dev"
}

// GracefulShutdown handles SIGINT and SIGTERM for clean server shutdown.
// It calls the provided shutdown function when a signal is received.
func GracefulShutdown(srv *http.Server, log *logrus.Entry) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}
}
