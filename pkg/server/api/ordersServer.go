package api

import (
	"WB_ZeroProject/internal/database"
	"encoding/json"
	"net/http"
	"sync"
)

const (
	Localhost   = "127.0.0.1"
	DefaultPort = 8080
)

type OrdersServer struct {
	DB   database.DBRepository
	Lock sync.Mutex
}

var _ ServerInterface = (*OrdersServer)(nil)

func NewOrdersServer(db database.DBRepository) *OrdersServer {
	return &OrdersServer{
		DB: db,
	}
}

func sendOrdersServerError(w http.ResponseWriter, code int, message string) {
	ordersError := Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(ordersError)
	if err != nil {
		return
	}
}

func (s *OrdersServer) GetOrders(w http.ResponseWriter, r *http.Request) {

}

func (s *OrdersServer) CreateOrder(w http.ResponseWriter, r *http.Request) {

}

func (s *OrdersServer) ShowOrderById(w http.ResponseWriter, r *http.Request, orderUid string) {

}
