package app

import (
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"token13/merchant-backend-go/internal/auth"
	"token13/merchant-backend-go/internal/config"
	applogger "token13/merchant-backend-go/internal/platform/logger"
	"token13/merchant-backend-go/internal/queue/rabbit"
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

	RabbitConn *amqp.Connection
	Publisher  *rabbit.Publisher
}

func Wire() (*Container, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := applogger.New(applogger.Options{Env: cfg.AppEnv})

	// DB
	db, err := postgres.Connect(cfg.DBDSN)
	if err != nil {
		return nil, err
	}
	authRepo := postgres.NewAuthRepo(db.SQL)

	// JWT
	jwtm := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer, 15*time.Minute)

	// Tron (stub for now)
	tronSvc := tron.NewStub()

	// RabbitMQ
	rabbitConn, err := rabbit.Connect(cfg.RabbitURL)
	if err != nil {
		return nil, err
	}

	publisher, err := rabbit.NewPublisher(rabbitConn, cfg.RabbitExchange)
	if err != nil {
		return nil, err
	}

	// Handlers
	authH := handlers.NewAuthHandler(authRepo, jwtm, tronSvc, publisher)
	api := NewAPI(log, authH)

	return &Container{
		Cfg:        cfg,
		Log:        log,
		DB:         db,
		AuthRepo:   authRepo,
		JWT:        jwtm,
		Tron:       tronSvc,
		RabbitConn: rabbitConn,
		Publisher:  publisher,
		API:        api,
	}, nil
}

func Addr(port int) string {
	return fmt.Sprintf(":%d", port)
}
