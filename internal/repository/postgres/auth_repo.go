package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type AuthRepo struct {
	db *sql.DB
}

func NewAuthRepo(db *sql.DB) *AuthRepo {
	return &AuthRepo{db: db}
}

// CreateMerchantAndUserTx creates merchant + user in a single DB transaction.
// Returns: merchantID (same input), status (usually "PENDING"), error
func (r *AuthRepo) CreateMerchantAndUserTx(
	ctx context.Context,
	merchantID []byte,
	name string,
	wallet string,
	email string,
	passwordHash string,
) ([]byte, string, error) {

	if len(merchantID) != 32 {
		return nil, "", fmt.Errorf("merchant_id must be 32 bytes, got %d", len(merchantID))
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback()

	// 1) merchants
	_, err = tx.ExecContext(ctx, `
		INSERT INTO merchants (merchant_id, name, wallet_address, status)
		VALUES ($1, $2, $3, 'PENDING')
	`, merchantID, name, wallet)
	if err != nil {
		return nil, "", mapSQLError(err)
	}

	// 2) users
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (merchant_id, email, password_hash, role, status)
		VALUES ($1, $2, $3, 'MERCHANT', 'ACTIVE')
	`, merchantID, email, passwordHash)
	if err != nil {
		return nil, "", mapSQLError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, "", err
	}

	return merchantID, "PENDING", nil
}

// GetUserByEmail returns primitive fields needed by handler.
// merchant_id can be NULL for ADMIN; in that case merchantID will be nil.
func (r *AuthRepo) GetUserByEmail(ctx context.Context, email string) (
	userUID string,
	emailOut string,
	passwordHash string,
	role string,
	status string,
	merchantID []byte,
	err error,
) {
	// merchant_id is BYTEA nullable: scanning into []byte works (nil if NULL)
	err = r.db.QueryRowContext(ctx, `
		SELECT user_uid, email, password_hash, role, status, merchant_id
		FROM users
		WHERE email = $1
	`, email).Scan(
		&userUID,
		&emailOut,
		&passwordHash,
		&role,
		&status,
		&merchantID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", "", "", nil, fmt.Errorf("user not found")
		}
		return "", "", "", "", "", nil, err
	}

	return userUID, emailOut, passwordHash, role, status, merchantID, nil
}

func mapSQLError(err error) error {
	msg := strings.ToLower(err.Error())

	// These constraint names must match your SQL indexes.
	// If your names differ, update these strings accordingly.
	switch {
	case strings.Contains(msg, "users_email_uidx"):
		return fmt.Errorf("email already exists")
	case strings.Contains(msg, "merchants_wallet_uidx"):
		return fmt.Errorf("wallet already exists")
	case strings.Contains(msg, "merchants_name_uidx"):
		return fmt.Errorf("merchant name already exists")
	case strings.Contains(msg, "duplicate key"):
		return fmt.Errorf("duplicate value")
	default:
		return err
	}
}
