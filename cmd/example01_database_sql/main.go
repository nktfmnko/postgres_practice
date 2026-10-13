package main

import (
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const dsn = "postgres://postgres:1234@localhost:5432/orders?sslmode=disable"

func main() {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("sql.Open", slog.Any("error", err))
		return
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		slog.Error("db.Ping", slog.Any("error", err))
		return
	}

	slog.Info("Соединение установлено")

	row := db.QueryRow("SELECT id, customer_name, status, total, created_at FROM orders LIMIT 1")

	var (
		id                   int64
		customerName, status string
		total                float64
		createdAt            time.Time
	)

	if err = row.Scan(&id, &customerName, &status, &total, &createdAt); err != nil {
		slog.Error("row.Scan", slog.Any("error", err))
		return
	}

	slog.Info("Получен заказ",
		slog.Int64("id", id),
		slog.String("customerName", customerName),
		slog.String("status", status),
		slog.Float64("total", total),
		slog.String("createdAt", createdAt.Format("2006-01-02 15:04:05")),
	)

	stats := db.Stats()
	slog.Info("Статистика пула (database/sql)",
		slog.Int("Открытых_соединений", stats.OpenConnections),
		slog.Int("используется_сейчас", stats.InUse),
		slog.Int("простаивает", stats.Idle),
		slog.Int64("ожидали_соединения", stats.WaitCount),
	)

}
