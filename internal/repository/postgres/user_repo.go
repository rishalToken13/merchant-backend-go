package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type User struct {
	ID           int64
	UserUID      string
	MerchantID   sql.NullInt64
	Email        string
	PasswordHash string
	Role         string
	Status       string
}

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, email, passwordHash, role string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	const q = `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, user_uid::text, merchant_id, email, password_hash, role, status
	`
	var u User
	err := r.db.QueryRowContext(ctx, q, email, passwordHash, role).
		Scan(&u.ID, &u.UserUID, &u.MerchantID, &u.Email, &u.PasswordHash, &u.Role, &u.Status)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	const q = `
		SELECT id, user_uid::text, merchant_id, email, password_hash, role, status
		FROM users
		WHERE email = $1
	`
	var u User
	err := r.db.QueryRowContext(ctx, q, email).
		Scan(&u.ID, &u.UserUID, &u.MerchantID, &u.Email, &u.PasswordHash, &u.Role, &u.Status)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find by email: %w", err)
	}
	return &u, nil
}
