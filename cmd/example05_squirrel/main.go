package main

import (
	"context"
	"log/slog"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dsn = "postgres://postgres:1234@localhost:5432/orders?sslmode=disable"

	tableName       = "orders"
	colID           = "id"
	colCustomerName = "customer_name"
	colStatus       = "status"
	colTotal        = "total"
	colCreatedAt    = "created_at"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

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
	if newID == 0 {
		return
	}
	demoSelectOne(ctx, pool, newID)
	demoSelectList(ctx, pool)
	demoUpdate(ctx, pool, newID)
	demoDelete(ctx, pool, newID)
}

func demoInsert(ctx context.Context, pool *pgxpool.Pool) int64 {
	insertSQL, insertArgs, err := psql.Insert(tableName).
		Columns(colCustomerName, colTotal).
		Values("Петр Петров", 250.00).
		Suffix("RETURNING " + colID).ToSql()
	if err != nil {
		slog.Error("ошибка построения INSERT", slog.Any("error", err))
		return 0
	}

	var newID int64
	err = pool.QueryRow(ctx, insertSQL, insertArgs...).Scan(&newID)
	if err != nil {
		slog.Error("ошибка INSERT", slog.Any("error", err))
		return 0
	}

	slog.Info("Создан заказ", slog.Int64("id", newID))

	return newID
}

func demoSelectOne(ctx context.Context, pool *pgxpool.Pool, id int64) {
	selectSQL, selectArgs, err := psql.
		Select(colID, colCustomerName, colStatus, colTotal, colCreatedAt).
		From(tableName).
		Where(sq.Eq{colID: id}).
		ToSql()
	if err != nil {
		slog.Error("ошибка построения SELECT", slog.Any("error", err))
		return
	}

	rows, err := pool.Query(ctx, selectSQL, selectArgs...)
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
	status := new("new")
	minTotal := new(100.00)

	qb := psql.
		Select(colID, colCustomerName, colStatus, colCreatedAt).
		From(tableName).
		OrderBy(colCreatedAt + "DESC")
	if status != nil {
		qb = qb.Where(sq.Eq{colStatus: *status})
	}
	if minTotal != nil {
		qb = qb.Where(sq.GtOrEq{colStatus: *status})
	}

	selectSQL, selectArgs, err := qb.ToSql()
	if err != nil {
		slog.Error("ошибка построения SELECT", slog.Any("error", err))
		return
	}

	rows, err := pool.Query(ctx, selectSQL, selectArgs...)
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
	updateSQL, updateArgs, err := psql.
		Update(tableName).
		Set(colStatus, "paid").
		Where(sq.Eq{colID: id}).
		ToSql()
	if err != nil {
		slog.Error("ошибка построения UPDATE", slog.Any("error", err))
		return
	}
	tag, err := pool.Exec(ctx, updateSQL, updateArgs...)
	if err != nil {
		slog.Error("ошибка UPDATE", slog.Any("error", err))
		return
	}
	slog.Info("UPDATE выполнен", slog.Int64("затронуто_строк", tag.RowsAffected()))
}

func demoDelete(ctx context.Context, pool *pgxpool.Pool, id int64) {
	deleteSQL, deleteArgs, err := psql.
		Delete(tableName).
		Where(sq.Eq{colID: id}).
		ToSql()
	if err != nil {
		slog.Error("ошибка построения DELETE", slog.Any("error", err))
		return
	}
	tag, err := pool.Exec(ctx, deleteSQL, deleteArgs...)
	if err != nil {
		slog.Error("ошибка DELETE", slog.Any("error", err))
		return
	}
	slog.Info("DELETE выполнен", slog.Int64("затронуто_строк", tag.RowsAffected()))
}
