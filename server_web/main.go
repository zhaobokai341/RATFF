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
	log      *logrus.Entry
	wsConn   *websocket.Conn
	wsConnMu sync.Mutex
)

// setupRouter configures the HTTP router with all web endpoints.
func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(rateLimitMiddleware())
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

	r.GET("/", handleRoot)
	r.GET("/ws", handleWebSocketRoot)
	r.GET("/api/clients", handleAPIClientsRoot)
	r.POST("/api/command", handleExecCommand)
	r.POST("/api/file/list", handleFileList)
	r.POST("/api/file/move", handleFileMove)
	r.POST("/api/file/delete", handleFileDelete)
	r.POST("/api/file/copy", handleFileCopy)

	r.GET("/:pathPassword", handlePathIndex)
	r.GET("/:pathPassword/", handlePathIndex)
	r.GET("/:pathPassword/index.html", handlePathIndex)
	r.GET("/:pathPassword/ws", handlePathWebSocket)
	r.GET("/:pathPassword/api/clients", handlePathAPIClients)
	r.POST("/:pathPassword/api/command", handleExecCommand)
	r.POST("/:pathPassword/api/file/list", handleFileList)
	r.POST("/:pathPassword/api/file/move", handleFileMove)
	r.POST("/:pathPassword/api/file/delete", handleFileDelete)
	r.POST("/:pathPassword/api/file/copy", handleFileCopy)

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
	log = shared.InitLogger("info", "text")
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
}
