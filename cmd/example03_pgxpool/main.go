package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const dsn = "postgres://postgres:1234@localhost:5432/orders?sslmode=disable"

type Order struct {
	ID           int64     `db:"id"`
	CustomerName string    `db:"customer_name"`
	Status       string    `db:"status"`
	Total        float64   `db:"total"`
	CreatedAt    time.Time `db:"created_at"`
}

func main() {
	ctx := context.Background()

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.Error("pgxpool.ParseConfig", slog.Any("error", err))
		return
	}

	config.MaxConns = 10
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		slog.Error("pgxpool.NewWithConfig", slog.Any("error", err))
		return
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		slog.Error("pool.Ping", slog.Any("error", err))
		return
	}

	rows, err := pool.Query(ctx, "SELECT id, customer_name, status, total, created_at FROM orders")
	if err != nil {
		slog.Error("pool.Query", slog.Any("error", err))
		return
	}

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[Order])
	if err != nil {
		slog.Error("pool.CollectRows", slog.Any("error", err))
		return
	}

	slog.Info("Получены заказы",
		slog.Int("количество", len(orders)),
	)

	for _, o := range orders {
		slog.Info("Заказ",
			slog.Int64("id", o.ID),
			slog.String("customerName", o.CustomerName),
			slog.String("status", o.Status),
			slog.Float64("total", o.Total),
			slog.String("createdAt", o.CreatedAt.Format("2006-01-02 15:04:05")),
		)
	}

	stats := pool.Stat()
	slog.Info("Статистика пула (pgxpool)",
		slog.Int("всего_соединений", int(stats.TotalConns())),
		slog.Int("простаивает", int(stats.IdleConns())),
		slog.Int("используется", int(stats.AcquiredConns())),
		slog.Int("максимум", int(stats.MaxConns())),
		slog.Int("конструируется_сейчас", int(stats.ConstructingConns())),
	)
}
