package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"token13/merchant-backend-go/internal/auth"
	"token13/merchant-backend-go/internal/repository/postgres"
)

type AuthHandler struct {
	Users *postgres.UserRepo
	JWT   *auth.JWTManager
}

func NewAuthHandler(users *postgres.UserRepo, jwtm *auth.JWTManager) *AuthHandler {
	return &AuthHandler{Users: users, JWT: jwtm}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.Users.FindByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.Users.Create(c.Request.Context(), req.Email, hash, "MERCHANT")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	var mid *int64
	if u.MerchantID.Valid {
		v := u.MerchantID.Int64
		mid = &v
	}

	token, err := h.JWT.Sign(u.UserUID, u.Role, mid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_uid": u.UserUID,
		"email":    u.Email,
		"role":     u.Role,
		"token":    token,
	})
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.Users.FindByEmail(c.Request.Context(), req.Email)
	if err != nil || u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if !auth.VerifyPassword(u.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	var mid *int64
	if u.MerchantID.Valid {
		v := u.MerchantID.Int64
		mid = &v
	}

	token, err := h.JWT.Sign(u.UserUID, u.Role, mid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_uid": u.UserUID,
		"email":    u.Email,
		"role":     u.Role,
		"token":    token,
	})
}
