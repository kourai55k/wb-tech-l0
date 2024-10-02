package cache

import (
	"WBTechL0/internal/db/repository"
	"WBTechL0/internal/models"
	"context"
	"fmt"
	"sync"
)

// Cache — структура для хранения кэша
type Cache struct {
	mu     sync.RWMutex // Для избежания гонки данных
	orders map[string]models.Order
}

// New — создание нового кэша с заданным TTL
func New() *Cache {
	return &Cache{
		orders: make(map[string]models.Order),
	}
}

// Set — добавляет заказ в кэш
func (c *Cache) Set(order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Сохраняем заказ в кэш
	c.orders[order.OrderUID] = order
}

// Get — извлекает заказ из кэша по его UID
func (c *Cache) Get(orderUID string) (*models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, found := c.orders[orderUID]
	if !found {
		return nil, false
	}
	return &order, found
}

// RestoreCacheFromDB загружает все заказы из базы данных в кэш
func (c *Cache) RestoreCacheFromDB(repo *repository.Repo) error {
	orders, err := repo.GetAllOrders(context.Background()) // Предполагаем, что эта функция возвращает все заказы
	if err != nil {
		return fmt.Errorf("failed to retrieve orders from database: %v", err)
	}

	// Загружаем заказы в кэш
	for _, order := range orders {
		c.Set(order) // Сохраняем каждый заказ в кэш
	}
	fmt.Printf("Cache restored with %d orders from the database\n", len(orders))
	return nil
}
