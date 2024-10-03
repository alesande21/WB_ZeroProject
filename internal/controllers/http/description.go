package http

import (
	entity2 "WB_ZeroProject/internal/entity"
	kafka2 "WB_ZeroProject/internal/kafka"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/invopop/yaml"
	"io"
	"log"
	"net/http"
	"os"
)

type ServerAddress struct {
	Localhost   string `json:"localhost" yaml:"localhost"`
	DefaultPort int    `json:"defaultPort" yaml:"defaultPort"`
	EnvAddress  string `env-required:"true" json:"envAddress" yaml:"envAddress" env:"SERVER_ADDRESS"`
}

func (a *ServerAddress) LoadConfigAddress(filePath string) error {
	_, err := os.Stat(filePath)
	if !(err == nil || !os.IsNotExist(err)) {
		return errors.New("конфиг для localhost и port не найден")
	}

	//if err != nil {
	//	if os.IsNotExist(err) {
	//		return errors.New("конфиг для localhost и port не найден")
	//	}
	//	return fmt.Errorf("ошибка проверки файла: %w", err)
	//}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("ошибка чтения конфига, %w", err)
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("ошибка чтения конфига, %w, %s", err, string(buf))
	}

	err = yaml.Unmarshal(buf, a)
	if err != nil {
		return fmt.Errorf("ошибка unmarshal, %w", err)
	}

	//УДАЛИТЬ
	a.Localhost = "127.0.0.1"
	a.DefaultPort = 8080
	a.EnvAddress = "127.0.0.1:8080"

	return nil
}

func (a *ServerAddress) UpdateEnvAddress() error {
	err := cleanenv.ReadEnv(a)
	if err != nil {
		return fmt.Errorf("ошибка updating env адреса сервера: %w", err)
	}
	return nil
}

type OrderServer struct {
	orderPlacer *kafka2.OrderPlacer
}

var _ ServerInterface = (*OrderServer)(nil)

func NewTenderServer(orderPlacer *kafka2.OrderPlacer) *OrderServer {
	return &OrderServer{
		orderPlacer: orderPlacer,
	}
}

type Error struct {
	Code    int32
	Message string
}

func sendErrorResponse(w http.ResponseWriter, code int, resp entity2.ErrorResponse) {
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

//func (o *OrderServer) GetOrders(w http.ResponseWriter, r *http.Request, params entity2.GetOrdersParams) {
//	limit := params.Limit
//	if limit == nil {
//		var valLimit entity2.PaginationLimit = 5
//		if params.Offset != nil {
//			limit = &valLimit
//		} else {
//			count, err := o.orderService.Repo.GetOrderCount(r.Context())
//			if err != nil {
//				sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка получения списка заказов"})
//				return
//			}
//			valLimit = entity2.PaginationLimit(count)
//			limit = &valLimit
//		}
//	}
//
//	offset := params.Offset
//	if offset == nil {
//		var defOffset entity2.PaginationOffset = 0
//		offset = &defOffset
//	}
//
//	orders, err := o.orderService.Repo.GetOrders(r.Context(), *limit, *offset)
//	if err != nil {
//		http.Error(w, "Ошибка получения списка заказов", http.StatusInternalServerError)
//		return
//	}
//
//	w.WriteHeader(http.StatusOK)
//	w.Header().Set("Content-Type", "application/json")
//	if err := json.NewEncoder(w).Encode(orders); err != nil {
//		http.Error(w, "Ошибка кодирования ответа", http.StatusBadRequest)
//	}
//}

func (os *OrderServer) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var newOrders []entity2.Order
	if err := json.NewDecoder(r.Body).Decode(&newOrders); err != nil {
		log.Println("Неверный формат для заказа!")
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат для заказа."})
		return
	}

	err := os.orderPlacer.CreateOrder(r.Context(), "orders.event.create", newOrders)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка создания заказа."})
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")

	// TODO: возможно добавить id ордеров или убрать вовсе
	msg := "Ордера приняты в обработку."
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}
}

func (os *OrderServer) GetOrderById(w http.ResponseWriter, r *http.Request, orderUid entity2.OrderId) {
	if orderUid == "" {
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат запроса или его параметры."})
		return
	}

	order, err := os.orderPlacer.GetOrder(r.Context(), "orders.event.getById", orderUid)
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Заказ не найден."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}
}

func (os OrderServer) GetApiPing(w http.ResponseWriter, r *http.Request) {
	res := "ok"
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		return
	}
}
