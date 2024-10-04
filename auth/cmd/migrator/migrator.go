package main

import (
	"auth/config"
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"log"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var migrationsPath string
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.Parse()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Ошибка при попытке спарсить .env файл в структуру: %v", err)
	}

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", cfg.DB.Username, cfg.DB.Password, cfg.DB.Address, cfg.DB.Port, cfg.DB.Name)

	m, err := migrate.New("file://"+migrationsPath, connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Нет миграций для применения")
			return
		}
		log.Fatalf("Ошибка применения миграции: %v", err)
	}

	log.Println("Миграции успешно применены")
}
