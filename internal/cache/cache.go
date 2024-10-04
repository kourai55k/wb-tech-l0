package cache

import (
	"WBTechL0/internal/db/repository"
	"WBTechL0/internal/models"
	"context"
	"sync"
)

// Cache — структура для хранения кэша
type Cache struct {
	mu     sync.RWMutex // Для избежания гонки данных
	orders map[string]models.Order
}

// New — создание нового кэша
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
func (c *Cache) RestoreCacheFromDB(repo *repository.Repo) (int, error) {
	orders, err := repo.GetAllOrders(context.Background())
	if err != nil {
		return 0, err
	}

	// Загружаем заказы в кэш
	for _, order := range orders {
		c.Set(order)
	}

	// Возвращаем количество записей, восстановленных из бд и ошибку
	return len(orders), nil
}
