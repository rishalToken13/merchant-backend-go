// internal/transport/http/handlers/auth.go
package handlers

import (
	"context"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"token13/merchant-backend-go/internal/auth"
	"token13/merchant-backend-go/internal/domain/ids"
)

// -------------------------
// Interfaces (repo + tron)
// -------------------------

type AuthRepo interface {
	CreateMerchantAndUserTx(
		ctx context.Context,
		merchantID []byte, // must be 32 bytes
		name string,
		wallet string,
		email string,
		passwordHash string,
	) (merchantIDOut []byte, status string, err error)

	GetUserByEmail(ctx context.Context, email string) (
		userUID string,
		emailOut string,
		passwordHash string,
		role string,
		status string,
		merchantID []byte,
		err error,
	)
}

type TronService interface {
	RegisterMerchant(ctx context.Context, merchantID []byte, walletAddress string) (txid string, err error)
}

// -------------------------
// Handler
// -------------------------

type AuthHandler struct {
	repo AuthRepo
	jwt  *auth.JWTManager
	tron TronService
}

func NewAuthHandler(repo AuthRepo, jwt *auth.JWTManager, tron TronService) *AuthHandler {
	return &AuthHandler{
		repo: repo,
		jwt:  jwt,
		tron: tron,
	}
}

// -------------------------
// DTOs
// -------------------------

type RegisterRequest struct {
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required,min=8"`
	WalletAddress string `json:"wallet_address" binding:"required"`
	Name          string `json:"name" binding:"required"`
}

type RegisterResponse struct {
	Merchant struct {
		MerchantID    string `json:"merchant_id"` // 0x... (hex of bytes32)
		Name          string `json:"name"`
		WalletAddress string `json:"wallet_address"`
		Status        string `json:"status"`
	} `json:"merchant"`
	Chain struct {
		Registered bool   `json:"registered"`
		Txid       string `json:"txid,omitempty"`
		Error      string `json:"error,omitempty"`
	} `json:"chain"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"` // seconds
	User        struct {
		UserUID    string `json:"user_uid"`
		Email      string `json:"email"`
		Role       string `json:"role"`
		MerchantID string `json:"merchant_id,omitempty"` // 0x... if present
	} `json:"user"`
}

// -------------------------
// Helpers
// -------------------------

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func normalizeWallet(s string) string {
	return strings.TrimSpace(s)
}

func bytes32ToHexOrEmpty(b []byte) string {
	if len(b) == 32 {
		return "0x" + hex.EncodeToString(b)
	}
	return ""
}

func isDisabled(status string) bool {
	return strings.ToUpper(strings.TrimSpace(status)) == "DISABLED"
}

// -------------------------
// Handlers
// -------------------------

// Register
// Flow: create merchant+user (tx) -> call chain -> respond with chain status
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Email = normalizeEmail(req.Email)
	req.WalletAddress = normalizeWallet(req.WalletAddress)
	req.Name = strings.TrimSpace(req.Name)

	merchantID, err := ids.NewBytes32()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate merchant_id"})
		return
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	_, status, err := h.repo.CreateMerchantAndUserTx(
		c.Request.Context(),
		merchantID,
		req.Name,
		req.WalletAddress,
		req.Email,
		string(passHash),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchantHex, _ := ids.Bytes32ToHex(merchantID)

	resp := RegisterResponse{}
	resp.Merchant.MerchantID = merchantHex
	resp.Merchant.Name = req.Name
	resp.Merchant.WalletAddress = req.WalletAddress
	resp.Merchant.Status = status

	// Chain call AFTER DB commit — keep DB even if chain fails
	if h.tron != nil {
		txid, err := h.tron.RegisterMerchant(c.Request.Context(), merchantID, req.WalletAddress)
		if err != nil {
			resp.Chain.Registered = false
			resp.Chain.Error = "skipped"
			//resp.Chain.Error = err.Error()
			c.JSON(http.StatusOK, resp)
			return
		}
		resp.Chain.Registered = true
		resp.Chain.Txid = txid
	} else {
		resp.Chain.Registered = false
		resp.Chain.Error = "tron service not configured"
	}

	c.JSON(http.StatusOK, resp)
}

// Login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Email = normalizeEmail(req.Email)

	userUID, emailOut, passHashDB, role, status, merchantID, err := h.repo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if isDisabled(status) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "account disabled"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passHashDB), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, expiresAt, err := h.jwt.Sign(userUID, emailOut, role, merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	merchantHex := bytes32ToHexOrEmpty(merchantID)

	resp := LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(expiresAt).Seconds()),
	}
	resp.User.UserUID = userUID
	resp.User.Email = emailOut
	resp.User.Role = role
	resp.User.MerchantID = merchantHex

	c.JSON(http.StatusOK, resp)
}
