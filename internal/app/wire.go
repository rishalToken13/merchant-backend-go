package app

import (
	"fmt"
	"log/slog"
	"time"

	"token13/merchant-backend-go/internal/auth"
	"token13/merchant-backend-go/internal/config"
	applogger "token13/merchant-backend-go/internal/platform/logger"
	"token13/merchant-backend-go/internal/repository/postgres"
	"token13/merchant-backend-go/internal/services/tron"
	"token13/merchant-backend-go/internal/transport/http/handlers"
)

type Container struct {
	Cfg *config.Config
	Log *slog.Logger
	API *API
	DB  *postgres.DB

	AuthRepo *postgres.AuthRepo
	JWT      *auth.JWTManager
	Tron     tron.Service
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

	authRepo := postgres.NewAuthRepo(db.SQL)
	jwtm := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer, 15*time.Minute)

	// stub for now
	tronSvc := tron.NewStub()

	authH := handlers.NewAuthHandler(authRepo, jwtm, tronSvc)
	api := NewAPI(log, authH)

	return &Container{
		Cfg:      cfg,
		Log:      log,
		DB:       db,
		AuthRepo: authRepo,
		JWT:      jwtm,
		Tron:     tronSvc,
		API:      api,
	}, nil
}

func Addr(port int) string {
	return fmt.Sprintf(":%d", port)
}
