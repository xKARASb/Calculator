package postgres

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type PostgresConfig struct {
	Username string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"123"`
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	DbName   string `env:"POSTGRES_DB" env-default:"users"`
}

type DB struct {
	Db *sqlx.DB
}

func New(config PostgresConfig) (*DB, error) {
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		config.Username, config.Password, config.Host, config.Port, config.DbName)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if _, err := db.Conn(context.Background()); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
