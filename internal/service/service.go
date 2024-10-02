package service

import (
	"WBTechL0/internal/cache"
	"WBTechL0/internal/db/repository"
	"WBTechL0/internal/models"
	"context"
	"github.com/go-playground/validator/v10"
	"log/slog"
)

type OrderService struct {
	sl    *slog.Logger
	cache *cache.Cache
	repo  *repository.Repo
}

func New(cache *cache.Cache, repo *repository.Repo, sl *slog.Logger) *OrderService {
	return &OrderService{cache: cache, repo: repo, sl: sl}
}

func (srv *OrderService) SaveOrder(order models.Order) {
	valid := ValidateOrder(order)
	if !valid {
		srv.sl.Error("Invalid order")
	} else {
		_ = srv.repo.SaveOrder(context.Background(), order)
		srv.cache.Set(order)
	}
}

func (srv *OrderService) GetOrder(uid string) *models.Order {
	order, found := srv.cache.Get(uid)
	if !found {
		order, _ = srv.repo.GetOrderByUID(context.Background(), uid)
	}

	return order
}

// ValidateOrder - Валидация модели
func ValidateOrder(order models.Order) bool {
	validate := validator.New()
	err := validate.Struct(order)

	if err != nil {
		return false
	}
	return true
}
