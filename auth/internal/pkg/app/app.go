package app

import (
	"auth/config"
	"auth/internal/app/services/auth"
	"auth/pkg/cache"
	"auth/pkg/db"
	"auth/pkg/logger"
	"auth/pkg/metrics"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type App struct {
	auth *auth.Service

	server *grpc.Server
	router *gin.Engine
}

func New(cfg *config.Configuration) (*App, error) {
	logger.Init(cfg.LoggerLevel)

	a := &App{}

	NewGRPC(a, cfg)

	err := cache.Init(cfg.Redis.Address, cfg.Redis.Port, cfg.Redis.Username, cfg.Redis.Password, cfg.Redis.ID)
	if err != nil {
		logger.Error("Ошибка при инициализации кэша", zap.Error(err))
	}

	err = db.Init(cfg.DB.Address, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password, cfg.DB.Name)
	if err != nil {
		logger.Fatal("Ошибка при инициализации БД", zap.Error(err))
	}

	go metrics.Init()

	return a, nil
}

func NewGRPC(a *App, cfg *config.Configuration) {
	a.server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(),
	)

	// обьявляем сервисы
	a.auth = auth.New(cfg.Auth.AccessSecret, cfg.Auth.RefreshSecret, cfg.Auth.Pepper)

	// регистрируем эндпоинты

}

func (a *App) RunGRPC(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("Ошибка при открытии listener: ", zap.Error(err))
	}

	err = a.server.Serve(lis)
	if err != nil {
		logger.Fatal("Ошибка при инициализации сервера: ", zap.Error(err))
		return err
	}

	return nil
}
