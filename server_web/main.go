package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var (
	log         *logrus.Entry
	asyncWriter *shared.AsyncWriter
	wsConn      *websocket.Conn
	wsConnMu    sync.Mutex
)

// setupRouter configures the HTTP router with all web endpoints.
func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(languageMiddleware())

	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*.html")

	r.GET("/login", handleLoginPage)
	r.POST("/login", handleLogin)
	r.GET("/logout", handleLogout)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	r.GET("/api/lang", handleGetLanguage)
	r.POST("/api/lang", handleSetLanguage)

	// Non-API endpoints with global rate limiting
	nonAPI := r.Group("")
	nonAPI.Use(rateLimitMiddleware())
	{
		nonAPI.GET("/", handleRoot)
		nonAPI.GET("/ws", handleWebSocketRoot)
		nonAPI.GET("/api/clients", handleAPIClientsRoot)
	}

	// API endpoints with per-client rate limiting
	api := r.Group("")
	api.Use(apiRateLimitMiddleware())
	{
		api.POST("/api/command", handleExecCommand)
		api.POST("/api/file/list", handleFileList)
		api.POST("/api/file/move", handleFileMove)
		api.POST("/api/file/delete", handleFileDelete)
		api.POST("/api/file/copy", handleFileCopy)
		api.POST("/api/file/upload", handleFileUpload)
		api.POST("/api/file/download", handleFileDownload)
		api.POST("/api/screen/capture", handleScreenCapture)
		api.POST("/api/public-ip", handleGetPublicIP)
		api.GET("/api/task/progress", handleTaskProgress)
		api.GET("/api/task/status", handleTaskStatus)
		api.GET("/api/file/download_result", handleDownloadResult)
	}

	// Path-prefixed routes with global rate limiting
	pathNonAPI := r.Group("/:pathPassword")
	pathNonAPI.Use(rateLimitMiddleware())
	{
		pathNonAPI.GET("/", handlePathIndex)
		pathNonAPI.GET("/index.html", handlePathIndex)
		pathNonAPI.GET("/ws", handlePathWebSocket)
		pathNonAPI.GET("/api/clients", handlePathAPIClients)
	}

	// Path-prefixed API endpoints with per-client rate limiting
	pathAPI := r.Group("/:pathPassword")
	pathAPI.Use(apiRateLimitMiddleware())
	{
		pathAPI.POST("/api/command", handleExecCommand)
		pathAPI.POST("/api/file/list", handleFileList)
		pathAPI.POST("/api/file/move", handleFileMove)
		pathAPI.POST("/api/file/delete", handleFileDelete)
		pathAPI.POST("/api/file/copy", handleFileCopy)
		pathAPI.POST("/api/file/upload", handleFileUpload)
		pathAPI.POST("/api/file/download", handleFileDownload)
		pathAPI.POST("/api/screen/capture", handleScreenCapture)
		pathAPI.POST("/api/public-ip", handleGetPublicIP)
		pathAPI.GET("/api/task/progress", handleTaskProgress)
		pathAPI.GET("/api/task/status", handleTaskStatus)
		pathAPI.GET("/api/file/download_result", handleDownloadResult)
	}

	return r
}

// handleRoot handles direct root access (no path prefix).
func handleRoot(c *gin.Context) {
	token, _ := c.Cookie("auth_token")
	if token == "" {
		c.Redirect(302, "/login")
		return
	}
	c.HTML(200, "index.html", gin.H{
		"title": T(c, "page_title"),
		"lang":  getLang(c),
	})
}

// handlePathIndex handles /<path>/ and /<path>/index.html.
func handlePathIndex(c *gin.Context) {
	pathPassword := c.Param("pathPassword")
	c.SetCookie("path_prefix", pathPassword, 3600, "/", "", cfg.CookieSecure, true)

	token, _ := c.Cookie("auth_token")
	if token == "" {
		c.Redirect(302, "/login")
		return
	}
	c.HTML(200, "index.html", gin.H{
		"title": T(c, "page_title"),
		"lang":  getLang(c),
	})
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
		log.WithError(err).Error("Server forced to shutdown")
	}
}

func main() {
	log, asyncWriter = shared.InitLoggerWithWriter("info", "text", true)
	loadConfig()

	if err := shared.LoadLanguagePacks("lang"); err != nil {
		log.WithError(err).Warn("Failed to load language packs")
	}

	router := setupRouter()

	srv := &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: router,
	}

	go gracefulShutdown(srv)

	log.Info("Starting Web server on " + srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.WithError(err).Fatal("Web server failed")
	}

	// Flush remaining logs before exit
	if asyncWriter != nil {
		asyncWriter.Close()
	}
}
