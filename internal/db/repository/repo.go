package repository

import (
	"WBTechL0/internal/models"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type Repo struct {
	pool *pgxpool.Pool
	sl   *slog.Logger
}

// New создаёт новый Repo
func New(pool *pgxpool.Pool, sl *slog.Logger) *Repo {
	return &Repo{pool: pool, sl: sl}
}

// SaveOrder Сохраняет ордер в базу данных
func (r *Repo) SaveOrder(ctx context.Context, order models.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.sl.Error("Failed to start transaction", err)
		return err
	}
	defer tx.Rollback(ctx)

	// Сохраняем доставку
	var deliveryID int
	deliveryQuery := `INSERT INTO deliveries (name, phone, zip, city, address, region, email) 
	                  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err = tx.QueryRow(ctx, deliveryQuery, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email).Scan(&deliveryID)
	if err != nil {
		r.sl.Error("Failed to insert delivery", err)
		return err
	}

	// Сохраняем платеж
	var paymentID int
	paymentQuery := `INSERT INTO payments (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) 
	                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
	err = tx.QueryRow(ctx, paymentQuery, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee).Scan(&paymentID)
	if err != nil {
		r.sl.Error("Failed to insert payment", err)
		return err
	}

	// Сохраняем заказ
	orderQuery := `INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard, delivery_id, payment_id) 
	               VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err = tx.Exec(ctx, orderQuery, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard, deliveryID, paymentID)
	if err != nil {
		r.sl.Error("Failed to insert order", err)
		return err
	}

	// Сохраняем товары
	itemQuery := `INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) 
	              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	for _, item := range order.Items {
		_, err := tx.Exec(ctx, itemQuery, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			r.sl.Error("Failed to insert item", err)
			return err
		}
	}

	// Коммит транзакции
	err = tx.Commit(ctx)
	if err != nil {
		r.sl.Error("Failed to commit transaction", err)
		return err
	}

	r.sl.Info("Order successfully saved", "order", order)
	return nil
}

// GetOrderByUID получает заказ по order_uid
func (r *Repo) GetOrderByUID(ctx context.Context, orderUID string) (*models.Order, error) {
	// Запрос для получения заказа и связанных данных (delivery, payment)
	orderQuery := `
	SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
	       d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
	       p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
	FROM orders o
	JOIN deliveries d ON o.delivery_id = d.id
	JOIN payments p ON o.payment_id = p.id
	WHERE o.order_uid = $1`

	var order models.Order
	var delivery models.Delivery
	var payment models.Payment

	err := r.pool.QueryRow(ctx, orderQuery, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
		&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
		&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDT,
		&payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.sl.Warn("Order not found", "order_uid", orderUID)
			return nil, errors.New("order not found")
		}
		r.sl.Error("Failed to retrieve order", err)
		return nil, err
	}

	order.Delivery = delivery
	order.Payment = payment

	// Запрос для получения товаров, связанных с заказом
	itemsQuery := `
	SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
	FROM items
	WHERE track_number = $1`

	rows, err := r.pool.Query(ctx, itemsQuery, order.TrackNumber)
	if err != nil {
		r.sl.Error("Failed to retrieve items", err)
		return nil, err
	}
	defer rows.Close()
	// Получаем список товаров в заказе
	var items []models.Item
	for rows.Next() {
		var item models.Item
		if err = rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status); err != nil {
			r.sl.Error("Failed to scan item", err)
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		r.sl.Error("Error during rows iteration", err)
		return nil, err
	}

	order.Items = items

	r.sl.Debug("Order retrieved successfully", "order_uid", orderUID, "order", order)
	return &order, nil
}

// GetAllOrders получает все заказы из базы данных
func (r *Repo) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	// Запрос для получения всех заказов и связанных данных (доставка, платеж)
	orderQuery := `
	SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
	       d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
	       p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
	FROM orders o
	JOIN deliveries d ON o.delivery_id = d.id
	JOIN payments p ON o.payment_id = p.id`

	rows, err := r.pool.Query(ctx, orderQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve orders from database: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var delivery models.Delivery
		var payment models.Payment

		// Сканируем данные заказа, доставки и платежа
		err = rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService,
			&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
			&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
			&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDT,
			&payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
		)
		if err != nil {
			r.sl.Error("Failed to scan order", err)
			return nil, err
		}

		order.Delivery = delivery
		order.Payment = payment

		// Получаем товары для каждого заказа
		itemsQuery := `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE track_number = $1`

		itemRows, err := r.pool.Query(ctx, itemsQuery, order.TrackNumber)
		if err != nil {
			r.sl.Error("Failed to retrieve items for order", "order_uid", order.OrderUID, err)
			return nil, err
		}

		var items []models.Item
		for itemRows.Next() {
			var item models.Item
			if err := itemRows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status); err != nil {
				r.sl.Error("Failed to scan item", err)
				itemRows.Close()
				return nil, err
			}
			items = append(items, item)
		}
		itemRows.Close()

		order.Items = items
		orders = append(orders, order)
	}

	if rows.Err() != nil {
		r.sl.Error("Error during rows iteration", err)
		return nil, err
	}

	r.sl.Debug("All orders retrieved successfully")
	return orders, nil
}
