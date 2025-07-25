package middleware

import (
	"fmt"
	"net/http"
	"paperplay/config"
	"paperplay/internal/model"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTService handles JWT operations
type JWTService struct {
	secretKey            string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	db                   *gorm.DB
}

// NewJWTService creates a new JWT service
func NewJWTService(config *config.Config, db *gorm.DB) *JWTService {
	return &JWTService{
		secretKey:            config.JWT.SecretKey,
		accessTokenDuration:  time.Duration(config.JWT.AccessTokenDuration) * time.Minute,
		refreshTokenDuration: time.Duration(config.JWT.RefreshTokenDuration) * 24 * time.Hour,
		db:                   db,
	}
}

// GenerateTokenPair generates both access and refresh tokens for a user
func (j *JWTService) GenerateTokenPair(user *model.User) (accessToken, refreshToken string, err error) {
	// Generate access token
	accessToken, err = j.GenerateAccessToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err = j.GenerateRefreshToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// GenerateAccessToken generates a new access token for the user
func (j *JWTService) GenerateAccessToken(user *model.User) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "paperplay",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// GenerateRefreshToken generates and stores a new refresh token for the user
func (j *JWTService) GenerateRefreshToken(user *model.User) (string, error) {
	// Clean up expired refresh tokens for this user
	if err := j.CleanupExpiredRefreshTokens(user.ID); err != nil {
		return "", fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	// Create new refresh token
	refreshToken := &model.RefreshToken{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(j.refreshTokenDuration),
	}

	if err := j.db.Create(refreshToken).Error; err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return refreshToken.Token, nil
}

// ValidateAccessToken validates and parses an access token
func (j *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateRefreshToken validates a refresh token and returns the associated user
func (j *JWTService) ValidateRefreshToken(tokenString string) (*model.User, error) {
	var refreshToken model.RefreshToken
	if err := j.db.Where("token = ?", tokenString).First(&refreshToken).Error; err != nil {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}

	if refreshToken.IsExpired() {
		// Clean up expired token
		j.db.Delete(&refreshToken)
		return nil, fmt.Errorf("refresh token has expired")
	}

	// Get associated user
	var user model.User
	if err := j.db.First(&user, "id = ?", refreshToken.UserID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (j *JWTService) RefreshAccessToken(refreshTokenString string) (string, error) {
	user, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	return j.GenerateAccessToken(user)
}

// RevokeRefreshToken revokes a refresh token
func (j *JWTService) RevokeRefreshToken(tokenString string) error {
	return j.db.Where("token = ?", tokenString).Delete(&model.RefreshToken{}).Error
}

// RevokeAllRefreshTokens revokes all refresh tokens for a user
func (j *JWTService) RevokeAllRefreshTokens(userID string) error {
	return j.db.Where("user_id = ?", userID).Delete(&model.RefreshToken{}).Error
}

// CleanupExpiredRefreshTokens removes expired refresh tokens for a user
func (j *JWTService) CleanupExpiredRefreshTokens(userID string) error {
	return j.db.Where("user_id = ? AND expires_at < ?", userID, time.Now()).Delete(&model.RefreshToken{}).Error
}

// AuthMiddleware creates a Gin middleware for JWT authentication
func (j *JWTService) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := j.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Get user from database to ensure user still exists
		var user model.User
		if err := j.db.First(&user, "id = ?", claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Set("user", &user)

		c.Next()
	}
}

// OptionalAuthMiddleware creates a middleware that optionally validates JWT
// Sets user context if valid token is provided, but doesn't abort if no token
func (j *JWTService) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		token := tokenParts[1]
		claims, err := j.ValidateAccessToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Get user from database
		var user model.User
		if err := j.db.First(&user, "id = ?", claims.UserID).Error; err != nil {
			c.Next()
			return
		}

		// Store user information in context
		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Set("user", &user)

		c.Next()
	}
}

// GetCurrentUser extracts the current user from Gin context
func GetCurrentUser(c *gin.Context) (*model.User, bool) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	user, ok := userInterface.(*model.User)
	return user, ok
}

// GetCurrentUserID extracts the current user ID from Gin context
func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// MustGetCurrentUser extracts the current user from context or panics
func MustGetCurrentUser(c *gin.Context) *model.User {
	user, exists := GetCurrentUser(c)
	if !exists {
		panic("user not found in context")
	}
	return user
}

// MustGetCurrentUserID extracts the current user ID from context or panics
func MustGetCurrentUserID(c *gin.Context) string {
	userID, exists := GetCurrentUserID(c)
	if !exists {
		panic("user_id not found in context")
	}
	return userID
}
