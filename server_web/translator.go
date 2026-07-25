package main

import (
	"RATFF/shared"

	"github.com/gin-gonic/gin"
)

const langCookieName = "app_lang"
const defaultLang = "zh"

// languageMiddleware reads the language from cookie and injects it into context.
func languageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang, err := c.Cookie(langCookieName)
		if err != nil || lang == "" {
			lang = defaultLang
		}
		c.Set("lang", lang)
		c.Next()
	}
}

// getLang retrieves the current language from context.
func getLang(c *gin.Context) string {
	if lang, exists := c.Get("lang"); exists {
		if l, ok := lang.(string); ok && l != "" {
			return l
		}
	}
	return defaultLang
}

// T translates a key using the language from the current request context.
func T(c *gin.Context, key string) string {
	return shared.T(getLang(c), key)
}

// Tf translates a key and formats it with arguments using the current request language.
func Tf(c *gin.Context, key string, args ...interface{}) string {
	return shared.Tf(getLang(c), key, args...)
}

// handleGetLanguage returns the current language setting.
func handleGetLanguage(c *gin.Context) {
	c.JSON(200, gin.H{"lang": getLang(c)})
}

// handleSetLanguage sets the language cookie.
func handleSetLanguage(c *gin.Context) {
	var req struct {
		Lang string `json:"lang" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid language code"})
		return
	}
	c.SetCookie(langCookieName, req.Lang, 365*86400, "/", "", cfg.CookieSecure, true)
	c.JSON(200, gin.H{"status": "ok"})
}
