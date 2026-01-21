package ids

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func NewBytes32() ([]byte, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func Bytes32ToHex(b []byte) (string, error) {
	if len(b) != 32 {
		return "", fmt.Errorf("expected 32 bytes, got %d", len(b))
	}
	return "0x" + hex.EncodeToString(b), nil
}
