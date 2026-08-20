package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// verifyPasswordWithAPI sends password to server_api for verification.
func verifyPasswordWithAPI(pathPassword, password string) (string, error) {
	var verifyURL string
	if pathPassword != "" {
		verifyURL = cfg.APIBaseURL + "/" + pathPassword + "/verify"
	} else {
		verifyURL = cfg.APIBaseURL + "/verify"
	}

	body := map[string]string{"password": password}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(verifyURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	return result.Token, nil
}

// handleLoginPage renders the login page.
func handleLoginPage(c *gin.Context) {
	c.HTML(200, "login.html", gin.H{
		"title": T(c, "login_title"),
		"lang":  getLang(c),
	})
}

// handleLogin verifies password via server_api and sets auth cookie.
func handleLogin(c *gin.Context) {
	password := c.PostForm("password")

	if password == "" {
		c.HTML(400, "login.html", gin.H{
			"title": T(c, "login_title"),
			"error": T(c, "login_error_required"),
			"lang":  getLang(c),
		})
		return
	}

	pathPassword, _ := c.Cookie("path_prefix")
	token, err := verifyPasswordWithAPI(pathPassword, password)
	if err != nil {
		log.WithError(err).Warn("Failed to connect to server_api")
		c.SetCookie("path_prefix", "", -1, "/", "", cfg.CookieSecure, true)
		c.HTML(401, "login.html", gin.H{
			"title": T(c, "login_title"),
			"error": T(c, "login_error_invalid"),
			"lang":  getLang(c),
		})
		return
	}

	c.SetCookie("auth_token", token, 3600, "/", "", cfg.CookieSecure, true)

	wsConnMu.Lock()
	if wsConn != nil {
		wsConn.Close()
	}
	newConn, err := connectWS(pathPassword)
	if err != nil {
		log.WithError(err).Warn("Failed to connect WebSocket")
	} else {
		wsConn = newConn
		go func() {
			_ = listenResponses(wsConn)
		}()
		log.Info("Connected to WebSocket server")
	}
	wsConnMu.Unlock()

	if pathPassword != "" {
		c.Redirect(302, "/"+pathPassword+"/")
	} else {
		c.Redirect(302, "/")
	}
}

// handleLogout clears the auth cookie.
func handleLogout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", cfg.CookieSecure, true)
	c.SetCookie("path_prefix", "", -1, "/", "", cfg.CookieSecure, true)
	c.Redirect(302, "/login")
}
