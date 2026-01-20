package app

import (
	"fmt"
	"log/slog"
	"time"

	"token13/merchant-backend-go/internal/auth"
	"token13/merchant-backend-go/internal/config"
	applogger "token13/merchant-backend-go/internal/platform/logger"
	"token13/merchant-backend-go/internal/repository/postgres"
	"token13/merchant-backend-go/internal/transport/http/handlers"
)

type Container struct {
	Cfg *config.Config
	Log *slog.Logger
	API *API
	DB  *postgres.DB

	UserRepo *postgres.UserRepo
	JWT      *auth.JWTManager
}

func Wire() (*Container, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := applogger.New(applogger.Options{Env: cfg.AppEnv})

	db, err := postgres.Connect(cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	userRepo := postgres.NewUserRepo(db.SQL)
	jwtm := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer, 15*time.Minute)

	authH := handlers.NewAuthHandler(userRepo, jwtm)
	api := NewAPI(log, authH)

	return &Container{
		Cfg:      cfg,
		Log:      log,
		DB:       db,
		UserRepo: userRepo,
		JWT:      jwtm,
		API:      api,
	}, nil
}

func Addr(port int) string {
	return fmt.Sprintf(":%d", port)
}