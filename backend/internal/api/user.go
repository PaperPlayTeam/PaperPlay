package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"paperplay/internal/middleware"
	"paperplay/internal/model"
	"paperplay/internal/service"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	db          *gorm.DB
	jwtService  *middleware.JWTService
	userService *service.UserService
	ethService  *service.EthereumService
	validator   *validator.Validate
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	db *gorm.DB,
	jwtService *middleware.JWTService,
	userService *service.UserService,
	ethService *service.EthereumService,
) *UserHandler {
	return &UserHandler{
		db:          db,
		jwtService:  jwtService,
		userService: userService,
		ethService:  ethService,
		validator:   validator.New(),
	}
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=6"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=50"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int         `json:"expires_in"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Register handles user registration
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Check if user already exists
	var existingUser model.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "user_exists",
			Message: "User with this email already exists",
		})
		return
	}

	// Create user
	user := &model.User{
		Email:       strings.ToLower(req.Email),
		DisplayName: req.DisplayName,
	}

	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "password_hash_error",
			Message: "Failed to process password",
		})
		return
	}

	// Generate Ethereum wallet if enabled
	if h.ethService != nil && h.ethService.IsEnabled() {
		address, privateKey, err := h.ethService.GenerateWallet()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "wallet_generation_error",
				Message: "Failed to generate Ethereum wallet",
				Details: err.Error(),
			})
			return
		}
		user.EthAddress = address
		user.EthPrivateKey = privateKey
	}

	// Save user to database
	if err := h.db.Create(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create user",
			Details: err.Error(),
		})
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.jwtService.GenerateTokenPair(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "token_generation_error",
			Message: "Failed to generate authentication tokens",
		})
		return
	}

	// Return response without password
	user.PasswordHash = ""

	c.JSON(http.StatusCreated, AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	})
}

// Login handles user login
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Find user by email
	var user model.User
	if err := h.db.Where("email = ?", strings.ToLower(req.Email)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid email or password",
		})
		return
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid email or password",
		})
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.jwtService.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "token_generation_error",
			Message: "Failed to generate authentication tokens",
		})
		return
	}

	// Return response without password
	user.PasswordHash = ""

	c.JSON(http.StatusOK, AuthResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	})
}

// RefreshToken handles token refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Generate new access token
	accessToken, err := h.jwtService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_refresh_token",
			Message: "Invalid or expired refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   15 * 60, // 15 minutes in seconds
	})
}

// Logout handles user logout
func (h *UserHandler) Logout(c *gin.Context) {
	userID := middleware.MustGetCurrentUserID(c)

	// Revoke all refresh tokens for this user
	if err := h.jwtService.RevokeAllRefreshTokens(userID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "logout_error",
			Message: "Failed to logout user",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Successfully logged out",
	})
}

// GetProfile returns current user profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	user := middleware.MustGetCurrentUser(c)

	// Load user progress statistics
	stats, err := h.userService.GetUserStats(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "stats_error",
			Message: "Failed to load user statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"stats": stats,
	})
}

// UpdateProfile updates user profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	user := middleware.MustGetCurrentUser(c)

	var updateReq struct {
		DisplayName string `json:"display_name" validate:"omitempty,min=2,max=50"`
		AvatarURL   string `json:"avatar_url" validate:"omitempty,url"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(updateReq); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Update user fields
	updates := make(map[string]interface{})
	if updateReq.DisplayName != "" {
		updates["display_name"] = updateReq.DisplayName
	}
	if updateReq.AvatarURL != "" {
		updates["avatar_url"] = updateReq.AvatarURL
	}

	if len(updates) > 0 {
		if err := h.db.Model(user).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "update_error",
				Message: "Failed to update profile",
			})
			return
		}
	}

	// Reload user
	if err := h.db.First(user, "id = ?", user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "reload_error",
			Message: "Failed to reload user data",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Profile updated successfully",
		Data:    user,
	})
}

// GetUserProgress returns user's learning progress
func (h *UserHandler) GetUserProgress(c *gin.Context) {
	userID := middleware.MustGetCurrentUserID(c)

	progress, err := h.userService.GetUserProgress(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "progress_error",
			Message: "Failed to load user progress",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "User progress loaded successfully",
		Data:    progress,
	})
}

// GetUserAchievements returns user's achievements
func (h *UserHandler) GetUserAchievements(c *gin.Context) {
	userID := middleware.MustGetCurrentUserID(c)

	achievements, err := h.userService.GetUserAchievements(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "achievements_error",
			Message: "Failed to load user achievements",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "User achievements loaded successfully",
		Data:    achievements,
	})
}
