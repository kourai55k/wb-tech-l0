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
	Sl    *slog.Logger
	Cache *cache.Cache
	Repo  *repository.Repo
}

func New(cache *cache.Cache, repo *repository.Repo, sl *slog.Logger) *OrderService {
	return &OrderService{Cache: cache, Repo: repo, Sl: sl}
}

func (srv *OrderService) SaveOrder(order models.Order) {
	valid := ValidateOrder(order)
	//valid := true
	if !valid {
		srv.Sl.Error("Invalid order", "order", order)
	} else {
		_ = srv.Repo.SaveOrder(context.Background(), order)
		srv.Cache.Set(order)
	}
}

func (srv *OrderService) GetOrder(uid string) *models.Order {
	order, found := srv.Cache.Get(uid)
	var err error
	if !found {
		order, err = srv.Repo.GetOrderByUID(context.Background(), uid)
	}
	if err != nil {
		srv.Sl.Error("Error in retrieving order by uid", "uid", uid)
	}

	srv.Sl.Info("Order retrieved successfully", "order_uid", uid, "order", order)
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
