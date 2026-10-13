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

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("pgxpool.ParseConfig", slog.Any("error", err))
		return
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		slog.Error("pool.Ping", slog.Any("error", err))
		return
	}

	newID := demoInsert(ctx, pool)
	demoSelectOne(ctx, pool, newID)
	demoSelectList(ctx, pool)
	demoUpdate(ctx, pool, newID)
	demoDelete(ctx, pool, newID)
}

func demoInsert(ctx context.Context, pool *pgxpool.Pool) int64 {
	var newID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO orders (customer_name, total) VALUES ($1, $2) RETURNING id`,
		"Иван иванов", 1000.00).Scan(&newID)
	if err != nil {
		slog.Error("ошибка INSERT", slog.Any("error", err))
		return 0
	}

	slog.Info("Создан заказ", slog.Int64("id", newID))

	return newID
}

func demoSelectOne(ctx context.Context, pool *pgxpool.Pool, id int64) {
	rows, err := pool.Query(ctx, `SELECT id, customer_name, status, total, created_at FROM orders WHERE id = $1`, id)
	if err != nil {
		slog.Error("ошибка SELECT", slog.Any("error", err))
		return
	}

	order, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Order])
	if err != nil {
		slog.Error("ошибка CollectOneRow", slog.Any("error", err))
		return
	}

	slog.Info("Получен заказ",
		slog.Int64("id", order.ID),
		slog.String("покупатель", order.CustomerName),
		slog.String("статус", order.Status),
		slog.Float64("цена", order.Total),
		slog.String("дата", order.CreatedAt.String()),
	)
}

func demoSelectList(ctx context.Context, pool *pgxpool.Pool) {
	rows, err := pool.Query(ctx, `SELECT id, customer_name, status, total, created_at FROM orders ORDER BY created_at DESC `)
	if err != nil {
		slog.Error("ошибка SELECT", slog.Any("error", err))
		return
	}

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[Order])
	if err != nil {
		slog.Error("ошибка CollectOneRow", slog.Any("error", err))
		return
	}

	for _, o := range orders {
		slog.Info("Заказ",
			slog.Int64("id", o.ID),
			slog.String("customerName", o.CustomerName),
			slog.String("status", o.Status),
			slog.Float64("total", o.Total),
			slog.String("createdAt", o.CreatedAt.Format("2006-01-02 15:04:05")),
		)
	}
}

func demoUpdate(ctx context.Context, pool *pgxpool.Pool, id int64) {
	tag, err := pool.Exec(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, "paid", id)
	if err != nil {
		slog.Error("ошибка UPDATE", slog.Any("error", err))
		return
	}
	slog.Info("UPDATE выполнен", slog.Int64("затронуто_строк", tag.RowsAffected()))
}

func demoDelete(ctx context.Context, pool *pgxpool.Pool, id int64) {
	tag, err := pool.Exec(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		slog.Error("ошибка DELETE", slog.Any("error", err))
		return
	}
	slog.Info("DELETE выполнен", slog.Int64("затронуто_строк", tag.RowsAffected()))
}
