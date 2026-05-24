package handler

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"nas-partner/backend/internal/audit"
	"nas-partner/backend/internal/config"
	"nas-partner/backend/internal/database"
	"nas-partner/backend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var cfg *config.Config

func SetConfig(c *config.Config) {
	cfg = c
}

func AuthStatus(c *gin.Context) {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"initialized": count > 0})
}

func AuthInit(c *gin.Context) {
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "already initialized"})
		return
	}

	var req model.InitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	_, err = database.DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", req.Username, string(hash))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"username": req.Username})
}

func AuthLogin(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ip := c.ClientIP()

	// Rate limiting: check consecutive failures in the last 15 minutes
	var lastSuccessTime string
	_ = database.DB.QueryRow(
		`SELECT COALESCE(MAX(created_at), '1970-01-01') FROM login_attempts WHERE username = ? AND success = 1`,
		req.Username,
	).Scan(&lastSuccessTime)

	var recentFailures int
	_ = database.DB.QueryRow(
		`SELECT COUNT(*) FROM login_attempts
		 WHERE username = ? AND success = 0
		 AND created_at > MAX(?, datetime('now', '-15 minutes'))`,
		req.Username, lastSuccessTime,
	).Scan(&recentFailures)

	if recentFailures >= 5 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "账户已被锁定，请15分钟后重试"})
		return
	}

	var user model.User
	err := database.DB.QueryRow("SELECT id, username, password_hash FROM users WHERE username = ?", req.Username).
		Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err == sql.ErrNoRows {
		database.DB.Exec(
			"INSERT INTO login_attempts (username, success, ip) VALUES (?, 0, ?)",
			req.Username, ip,
		)
		audit.Log(req.Username, "login_failed", "用户不存在", ip)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		database.DB.Exec(
			"INSERT INTO login_attempts (username, success, ip) VALUES (?, 0, ?)",
			req.Username, ip,
		)
		audit.Log(req.Username, "login_failed", "密码错误", ip)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	database.DB.Exec(
		"INSERT INTO login_attempts (username, success, ip) VALUES (?, 1, ?)",
		user.Username, ip,
	)
	audit.Log(user.Username, "login_success", "登录成功", ip)

	token, err := generateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, model.AuthResponse{Token: token, Username: user.Username})
}

func AuthRefresh(c *gin.Context) {
	username, _ := c.Get("username")

	token, err := generateToken(username.(string))
	if err != nil {
		log.Printf("token refresh failed for %s: %v", username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, model.AuthResponse{Token: token, Username: username.(string)})
}

func generateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}
