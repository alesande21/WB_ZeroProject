package http

import (
	entity2 "WB_ZeroProject/internal/entity"
	kafka2 "WB_ZeroProject/internal/kafka"
	"encoding/json"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/invopop/yaml"
	log2 "github.com/sirupsen/logrus"
	"io"
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
		return fmt.Errorf("-> os.Stat: файла не существует %s: %w", filePath, err)
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("-> os.OpenFile: ошибка при открытии файла %s: %w", filePath, err)

	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("-> os.OpenFile: ошибка при чтении файла %s: %w", filePath, err)
	}

	err = yaml.Unmarshal(buf, a)
	if err != nil {
		return fmt.Errorf("->  yaml.Unmarshal: ошибка при конвертации: %w", err)
	}

	return nil
}

func (a *ServerAddress) UpdateEnvAddress() error {
	err := cleanenv.ReadEnv(a)
	if err != nil {
		return fmt.Errorf("-> cleanenv.ReadEnv: ошибка загрузки параметров из переменных окружения: %w", err)
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
		log2.Infof("sendErrorResponse: ошибка при формировании ответа ошибки %s: %s", resp, err.Error())
	}
}

func (os OrderServer) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var newOrders []entity2.Order
	if err := json.NewDecoder(r.Body).Decode(&newOrders); err != nil {
		log2.Errorf("CreateOrder-> json.NewDecoder: неверный формат для заказа: %s", err.Error())
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат для заказа."})
		return
	}

	err := os.orderPlacer.CreateOrder("orders.event.request.create", newOrders)
	if err != nil {
		log2.Errorf("CreateOrder-> os.orderPlacer.CreateOrder%s", err.Error())
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка создания заказа."})
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	msg := "Ордера приняты в обработку."
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		log2.Errorf("CreateOrder-> json.NewEncoder: ошибка при кодирования овета: %s", err.Error())
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}
}

func (os OrderServer) GetOrderById(w http.ResponseWriter, r *http.Request, orderUid entity2.OrderId) {
	if orderUid == "" {
		log2.Error("GetOrderById: неверный формат запроса или его параметры: оrderID пустой")
		sendErrorResponse(w, http.StatusBadRequest, entity2.ErrorResponse{Reason: "Неверный формат запроса или его параметры. Не указан OrderId."})
		return
	}

	order, err := os.orderPlacer.GetOrder("orders.event.request.getByID", orderUid)
	if err != nil {
		log2.Errorf("GetOrderById-> os.orderPlacer.GetOrder%s", err.Error())
		sendErrorResponse(w, http.StatusNotFound, entity2.ErrorResponse{Reason: "Заказ не найден."})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log2.Errorf("GetOrderById-> json.NewEncoder: ошибка при кодирования овета: %s", err.Error())
		sendErrorResponse(w, http.StatusInternalServerError, entity2.ErrorResponse{Reason: "Ошибка кодирования ответа."})
	}
}

func (os OrderServer) GetApiPing(w http.ResponseWriter, r *http.Request) {
	res := "ok"
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log2.Errorf("GetApiPing-> json.NewEncoder: ошибка при кодирования овета: %s", err.Error())
	}
}
