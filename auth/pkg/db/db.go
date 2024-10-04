package db

import (
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var Conn *sqlx.DB

func Init(Address string, Port string, Username string, Password string, Name string) error {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", Username, Password, Address, Port, Name)

	var err error
	Conn, err = sqlx.Open("pgx", connStr)
	if err != nil {
		return err
	}

	// Настройка пула соединений
	Conn.SetMaxOpenConns(100)
	Conn.SetMaxIdleConns(50)
	Conn.SetConnMaxLifetime(time.Hour)

	// Проверка подключения к базе данных
	err = Conn.Ping()
	if err != nil {
		return err
	}

	return nil
}
