package kafka

import (
	config2 "WB_ZeroProject/internal/config"
	entity2 "WB_ZeroProject/internal/entity"
	"WB_ZeroProject/internal/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

type OrderPlacer struct {
	producer   *kafka.Producer
	topic      string
	deliveryCh chan kafka.Event
}

type event struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

type eventCreate struct {
	Type  string          `json:"type"`
	Value []entity2.Order `json:"value"`
}

type eventGet struct {
	Type          string          `json:"type"`
	Value         entity2.OrderId `json:"value"`
	CorrelationID string          `json:"correlation_id"`
}

func NewOrderPlacer(conf *config2.ConfigKafka) (*OrderPlacer, error) {
	configMap := kafka.ConfigMap{
		"bootstrap.servers": conf.URL,
		"client.id":         "orderPlacer",
		"acks":              "all",
	}

	client, err := kafka.NewProducer(&configMap)
	if err != nil {
		return nil, fmt.Errorf("kafka.NewProducer %w", err)
	}

	return &OrderPlacer{
		producer:   client,
		topic:      conf.Topic,
		deliveryCh: make(chan kafka.Event, 10000),
	}, nil
}

func (op *OrderPlacer) PlaceOrder(orderType string, size int) error {
	var (
		format  = fmt.Sprintf("%s - %d", orderType, size)
		payload = []byte(format)
	)

	err := op.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &op.topic,
			Partition: kafka.PartitionAny,
		},
		Value: payload,
	},
		op.deliveryCh,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	<-op.deliveryCh

	return nil
}

func (op *OrderPlacer) CreateOrder(ctx context.Context, msgType string, orders []entity2.Order) error {

	var b bytes.Buffer

	evt := eventCreate{
		Type:  msgType,
		Value: orders,
	}

	err := json.NewEncoder(&b).Encode(evt)

	if err != nil {
		return fmt.Errorf("ошибка кодироования сообщения event: %w", err)
	}

	err = op.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &op.topic,
			Partition: kafka.PartitionAny,
		},
		Value: b.Bytes(),
	},
		op.deliveryCh,
	)

	if err != nil {
		return fmt.Errorf("ошибка при producer.Produce: %w", err)
	}

	return nil
}

// TODO: понять как вернуть ордер из кафка. Использовать ли мапу?
func (op *OrderPlacer) GetOrder(ctx context.Context, msgType string, orderId entity2.OrderId) (*entity2.Order, error) {

	var b bytes.Buffer
	correlationID, err := utils.GenerateUUIDV7()
	if err != nil {
		return nil, fmt.Errorf("ошибка при генерации uuid -> utils.GenerateUUIDV7: %w", err)
	}

	evt := eventGet{
		Type:          msgType,
		Value:         orderId,
		CorrelationID: correlationID,
	}

	err = json.NewEncoder(&b).Encode(evt)

	if err != nil {
		return nil, fmt.Errorf("ошибка кодироования сообщения event -> NewEncoder.Encode: %w", err)
	}

	err = op.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &op.topic,
			Partition: kafka.PartitionAny,
		},
		Value: b.Bytes(),
	},
		op.deliveryCh,
	)

	return nil, nil
}

func (op *OrderPlacer) Close() {
	op.producer.Close()
}
