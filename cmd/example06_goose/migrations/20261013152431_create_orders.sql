-- +goose Up
CREATE TABLE IF NOT EXISTS orders
(
    id            SERIAL PRIMARY KEY,
    customer_name TEXT           NOT NULL,
    status        TEXT           NOT NULL DEFAULT 'new',
    total         NUMERIC(10, 2) NOT NULL DEFAULT 0,
    created_at    TIMESTAMP      NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS orders;
