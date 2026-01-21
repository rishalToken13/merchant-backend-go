package tron

import (
	"context"
	"fmt"
)

type Service interface {
	RegisterMerchant(ctx context.Context, merchantID []byte, walletAddress string) (txid string, err error)
}

type Stub struct{}

func NewStub() *Stub { return &Stub{} }

func (s *Stub) RegisterMerchant(ctx context.Context, merchantID []byte, walletAddress string) (string, error) {
	return "", fmt.Errorf("tron register not implemented yet")
}
