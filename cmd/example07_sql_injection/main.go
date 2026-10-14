package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderResult struct {
	ID           int64   `db:"id"`
	CustomerName string  `db:"customer_name"`
	Status       string  `db:"status"`
	Total        float64 `db:"total"`
}

const dsn = "postgres://postgres:1234@localhost:5432/orders?sslmode=disable"

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

	//2026/10/14 15:27:50 ERROR Ошибка запроса error="ОШИБКА: в подготовленный
	//оператор нельзя вставить несколько команд (SQLSTATE 42601)"
	demoDeleteBlocked(ctx, pool)

	//demoDeleteViaExec(ctx, pool)

	// без $ проходит, с ним - нет
	//demoUnionInjection(ctx, pool)
}

func demoDeleteBlocked(ctx context.Context, pool *pgxpool.Pool) {
	slog.Info("DELETE-инъенция через Query (extended protocol)")

	countOrders(ctx, pool, "ДО атаки")
	searchOrdersViaQuery(ctx, pool, "Alice")

	slog.Info("АТАКА DELETE FROM orders")

	searchOrdersViaQuery(ctx, pool, "'; DELETE FROM orders; --")
	countOrders(ctx, pool, "ПОСЛЕ атаки")
}

func demoDeleteViaExec(ctx context.Context, pool *pgxpool.Pool) {
	slog.Info("DELETE-инъенция через Exec (simple protocol)")
	countOrders(ctx, pool, "ДО атаки")

	updateStatusUnsafe(ctx, pool, "Alice")
	countOrders(ctx, pool, "Alice")

	slog.Info("АТАКА DELETE FROM orders")
	updateStatusUnsafe(ctx, pool, "'; DELETE FROM orders; --")
	countOrders(ctx, pool, "ПОСЛЕ атаки")
}

func demoUnionInjection(ctx context.Context, pool *pgxpool.Pool) {
	searchOrdersViaQuery(ctx, pool, "Alice")
	payload := "' UNION SELECT id, delivery_address, phone, total FROM orders --"
	searchOrdersViaQuery(ctx, pool, payload)
}

func updateStatusUnsafe(ctx context.Context, pool *pgxpool.Pool, customerName string) {
	query := "UPDATE orders SET status = 'viewed' WHERE customer_name = '" + customerName + "'"
	_, err := pool.Exec(ctx, query)
	if err != nil {
		slog.Error("Ошибка Exec", slog.Any("error", err))
	}
}

// так никогда не делаем, только через $ передаем параметры(тут только для примера)
func searchOrdersViaQuery(ctx context.Context, pool *pgxpool.Pool, customerName string) {
	query := "SELECT id, customer_name, status, total FROM orders WHERE customer_name = '" + customerName + "'"
	rows, err := pool.Query(ctx, query)
	if err != nil {
		slog.Error("Ошибка запроса", slog.Any("error", err))
		return
	}

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[OrderResult])
	if err != nil {
		slog.Error("Ошибка CollectRows", slog.Any("error", err))
	}

	for _, o := range orders {
		slog.Info("Результат",
			slog.Int64("id", o.ID),
			slog.String("customerName", o.CustomerName),
			slog.String("status", o.Status),
			slog.Float64("total", o.Total),
		)
	}
}

func countOrders(ctx context.Context, pool *pgxpool.Pool, label string) {
	var count int
	err := pool.QueryRow(ctx, "SELECT count(*) FROM orders").Scan(&count)
	if err != nil {
		slog.Error("Ошибка подсчета", slog.Any("error", err))
		return
	}
	slog.Info(fmt.Sprintf("Заказов в таблице (%s): %d", label, count))
}
