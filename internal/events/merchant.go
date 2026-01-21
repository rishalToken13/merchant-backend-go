package events

import "time"

type MerchantCreated struct {
	MerchantID    string    `json:"merchant_id"`    // 0x... bytes32
	WalletAddress string    `json:"wallet_address"` // Tron base58
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	CreatedAt     time.Time `json:"created_at"`
}
