package service

import (
	entity2 "WB_ZeroProject/internal/entity"
	"context"
	"log"
)

type OrderRepo interface {
	CreateOrder(ctx context.Context, newOrders []entity2.Order) ([]entity2.Order, error)
	GetOrders(ctx context.Context, limit entity2.PaginationOffset, offset entity2.PaginationOffset) ([]entity2.Order, error)
	GetOrderByIdFromDb(ctx context.Context, orderId entity2.OrderId) (*entity2.Order, error)
	GetOrderByIdFromCache(orderId entity2.OrderId) (*entity2.Order, error)
	GetOrderCount(ctx context.Context) (int, error)
	UpdateCache(ctx context.Context)
	GetOrderItems(ctx context.Context, orderUid string) ([]entity2.Item, error)
	Ping() error
}

type OrderService struct {
	Repo OrderRepo
}

func NewOrderService(repo OrderRepo) *OrderService {
	return &OrderService{Repo: repo}
}

func (s *OrderService) GetOrderById(ctx context.Context, orderId entity2.OrderId) (*entity2.Order, error) {
	order, err := s.Repo.GetOrderByIdFromCache(orderId)
	if err != nil {
		order, err = s.Repo.GetOrderByIdFromDb(ctx, orderId)
		if err != nil {
			return nil, err
		}
		log.Printf("Данные для заказа %s взяты из базы данных. Идет обновление кеша.", orderId)
		s.Repo.UpdateCache(ctx)
		return order, nil
	}
	log.Printf("Данные для заказа %s взяты из кеша.", orderId)
	return order, nil
}
