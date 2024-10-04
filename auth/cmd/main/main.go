package main

import (
	"auth/config"
	"auth/internal/pkg/app"
	"log"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Ошибка при попытке спарсить .env файл в структуру: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = application.RunGRPC(cfg.GRPC.Port)
	if err != nil {
		log.Fatal(err)
	}
}
