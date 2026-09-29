package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"pos-app/internal/platform/config"
)

func Connect(cfg config.Config) *pgxpool.Pool {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	return pool
}