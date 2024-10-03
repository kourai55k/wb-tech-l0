package db

import (
	"WBTechL0/internal/config"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectToDB подключается к базе данных, используя данные из конфига
func ConnectToDB(dbCfg config.Database) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		dbCfg.User,
		dbCfg.Password,
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.DBname,
	)
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	// Check if the database is accessible by performing a simple query
	err = pool.QueryRow(context.Background(), "SELECT 1").Scan(new(int))
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to access the database: %v", err)
	}

	return pool, nil
}

// CreateTables создает таблицы для хранения данных о заказах
func CreateTables(pool *pgxpool.Pool) error {
	createOrdersTable := `
	CREATE TABLE IF NOT EXISTS orders (
		order_uid VARCHAR(255) PRIMARY KEY,
		track_number VARCHAR(255) NOT NULL UNIQUE,
		entry VARCHAR(255) NOT NULL,
		locale VARCHAR(10) NOT NULL,
		internal_signature VARCHAR(255),
		customer_id VARCHAR(255) NOT NULL,
		delivery_service VARCHAR(255) NOT NULL,
		shardkey VARCHAR(255) NOT NULL,
		sm_id INT NOT NULL CHECK (sm_id >= 1),
		date_created TIMESTAMP NOT NULL,
		oof_shard VARCHAR(255) NOT NULL,
		delivery_id INT REFERENCES deliveries(id),  -- Внешний ключ на таблицу deliveries
		payment_id INT REFERENCES payments(id)       -- Внешний ключ на таблицу payments
	);`

	createDeliveriesTable := `
	CREATE TABLE IF NOT EXISTS deliveries (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		phone VARCHAR(20) NOT NULL,
		zip VARCHAR(7) NOT NULL,
		city VARCHAR(255) NOT NULL,
		address VARCHAR(255) NOT NULL,
		region VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL
	);`

	createPaymentsTable := `
	CREATE TABLE IF NOT EXISTS payments (
		id SERIAL PRIMARY KEY,
		transaction VARCHAR(255) NOT NULL,
		request_id VARCHAR(255),
		currency VARCHAR(10) NOT NULL,
		provider VARCHAR(50) NOT NULL,
		amount INT NOT NULL CHECK (amount >= 0),
		payment_dt BIGINT NOT NULL,
		bank VARCHAR(255) NOT NULL,
		delivery_cost INT CHECK (delivery_cost >= 0),
		goods_total INT CHECK (goods_total >= 0),
		custom_fee INT CHECK (custom_fee >= 0)
	);`

	createItemsTable := `
	CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		chrt_id INT NOT NULL,
		track_number VARCHAR(255) REFERENCES orders(track_number), -- Внешний ключ на таблицу orders
		price INT NOT NULL CHECK (price >= 0),
		rid VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		sale INT CHECK (sale >= 0),
		size VARCHAR(50) NOT NULL,
		total_price INT NOT NULL CHECK (total_price >= 0),
		nm_id INT NOT NULL,
		brand VARCHAR(255) NOT NULL,
		status INT NOT NULL
	);`

	// Выполняем команды для создания таблиц
	commands := []string{createDeliveriesTable, createPaymentsTable, createOrdersTable, createItemsTable}
	for _, cmd := range commands {
		if _, err := pool.Exec(context.Background(), cmd); err != nil {
			return err
		}
	}
	return nil
}
