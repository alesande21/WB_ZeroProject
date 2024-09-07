package api

import (
	"WB_ZeroProject/internal/database"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

const (
	Localhost   = "127.0.0.1"
	DefaultPort = 8080
)

type OrdersServer struct {
	DB database.DBRepository
	sync.Mutex
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
	s.Lock()
	orders, err := s.DB.Query("SELECT * FROM DataOrders.orders LIMIT 10")
	s.Unlock()

	if err != nil {
		return
	}

	response, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (s *OrdersServer) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var newOrder Order
	if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
		sendOrdersServerError(w, http.StatusBadRequest, "Неверный формат для нового заказа(newOrder)")
		return
	}
	s.Lock()
	defer s.Unlock()

	order, err := s.DB.QueryRow(s.GetStringForCreationOrder(), newOrder.OrderUid, newOrder.TrackNumber,
		newOrder.Entry, newOrder.Delivery.Id, newOrder.Payment.Transaction, newOrder.Items, newOrder.Locale,
		newOrder.Locale, newOrder.InternalSignature, newOrder.CustomerId, newOrder.DeliveryService,
		newOrder.Shardkey, newOrder.SmId, newOrder.DateCreated, newOrder.OofShard)

	if err != nil {
		log.Println("Failed to insert row:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	resp := "Новый заказ создан под UID: " + order.OrderUid
	json.NewEncoder(w).Encode(resp)
}

func (s *OrdersServer) GetStringForCreationOrder() string {
	return "INSERT INTO DataOrders.orders(order_uid, track_number, entry, delivery," +
		"payment, items, locale, internal_signature, customer_id, delivery_service," +
		"shardkey, sm_id,  date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6," +
		"$7, $8, $9, $10, $11, $12, $13, $14) RETURNING order_uid, track_number," +
		"entry, delivery, payment, items, locale, internal_signature, customer_id," +
		"delivery_service, shardkey, sm_id, date_created, oof_shard"
}

func (s *OrdersServer) ShowOrderById(w http.ResponseWriter, r *http.Request, orderUid string) {

}
