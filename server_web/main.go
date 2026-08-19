package main

import (
	"embed"
	"fmt"
	"net/http"
	"sync"

	"RATFF/shared"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

//go:embed lang/*.json
var langFS embed.FS

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
	nonAPI.Use(shared.GlobalRateLimitMiddleware())
	{
		nonAPI.GET("/", handleRoot)
		nonAPI.GET("/ws", handleWebSocketRoot)
		nonAPI.GET("/api/clients", handleAPIClientsRoot)
	}

	// API endpoints with per-client rate limiting
	api := r.Group("")
	api.Use(shared.PerClientRateLimitMiddleware(globalPerClientLimiter, func(c *gin.Context) string {
		token, _ := c.Cookie("auth_token")
		return token
	}))
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
		api.POST("/api/service/update", handleServiceUpdate)
		api.GET("/api/task/progress", handleTaskProgress)
		api.GET("/api/task/status", handleTaskStatus)
		api.GET("/api/file/download_result", handleDownloadResult)
	}

	// Path-prefixed routes with global rate limiting
	pathNonAPI := r.Group("/:pathPassword")
	pathNonAPI.Use(shared.GlobalRateLimitMiddleware())
	{
		pathNonAPI.GET("/", handlePathIndex)
		pathNonAPI.GET("/index.html", handlePathIndex)
		pathNonAPI.GET("/ws", handlePathWebSocket)
		pathNonAPI.GET("/api/clients", handlePathAPIClients)
	}

	// Path-prefixed API endpoints with per-client rate limiting
	pathAPI := r.Group("/:pathPassword")
	pathAPI.Use(shared.PerClientRateLimitMiddleware(globalPerClientLimiter, func(c *gin.Context) string {
		token, _ := c.Cookie("auth_token")
		return token
	}))
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
		pathAPI.POST("/api/service/update", handleServiceUpdate)
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

func main() {
	fmt.Println(shared.SelectLogo())
	log, asyncWriter = shared.InitLoggerWithWriter("info", "text", true)
	loadConfig()

	if err := shared.LoadEmbeddedLanguagePacks(langFS, "lang"); err != nil {
		log.WithError(err).Warn("Failed to load language packs")
	}

	router := setupRouter()

	srv := &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: router,
	}

	go shared.GracefulShutdown(srv, log)

	log.Info("Starting Web server on " + srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.WithError(err).Fatal("Web server failed")
	}

	// Flush remaining logs before exit
	if asyncWriter != nil {
		asyncWriter.Close()
	}
}
