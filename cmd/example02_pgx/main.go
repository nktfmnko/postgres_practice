package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
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

	conn, err := pgx.Connect(ctx, dsn) // одно соединение
	if err != nil {
		slog.Error("pgx.Connect", slog.Any("error", err))
		return
	}
	defer conn.Close(ctx)

	//demoCollectRows(ctx, conn)

	demoConnBusy(ctx, conn)
}

func demoCollectRows(ctx context.Context, conn *pgx.Conn) {
	rows, err := conn.Query(ctx, "SELECT id, customer_name, status, total, created_at FROM orders")
	if err != nil {
		slog.Error("Query", slog.Any("error", err))
		return
	}

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[Order])
	if err != nil {
		slog.Error("CollectRows", slog.Any("error", err))
		return
	}

	slog.Info("Получены заказы через CollectRows",
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
}

func demoConnBusy(ctx context.Context, conn *pgx.Conn) {
	slog.Info("Демо: конкурентный доступ к pgx.Conn")

	var wg sync.WaitGroup

	for i := range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var id int64
			err := conn.QueryRow(ctx, "SELECT id FROM orders LIMIT 1").Scan(&id)
			if err != nil {
				slog.Error("Горутина: ошибка",
					slog.Int("горутина", i),
					slog.Any("error", err),
				)
				return
			}

			slog.Info("Горутина:успех",
				slog.Int("горутина", i),
				slog.Int64("id", id))
		}()
	}

	wg.Wait()
}
