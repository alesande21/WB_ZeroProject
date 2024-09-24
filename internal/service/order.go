package service

import (
	entity2 "WB_ZeroProject/internal/entity"
	"context"
)

type OrderRepo interface {
	CreateOrder(ctx context.Context, newOrders []entity2.Order) ([]entity2.Order, error)
	GetOrders(ctx context.Context, limit entity2.PaginationOffset, offset entity2.PaginationOffset) ([]entity2.Order, error)
	GetOrderById(ctx context.Context, orderId entity2.OrderId) (*entity2.Order, error)
	GetOrderCount(ctx context.Context) (int, error)
	Ping() error
}

type OrderService struct {
	Repo OrderRepo
}

func NewOrderService(repo OrderRepo) *OrderService {
	return &OrderService{Repo: repo}
}
