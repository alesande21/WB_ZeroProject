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
	"sync"
	"time"
)

type OrderPlacer struct {
	producer    *kafka.Producer
	consumer    *kafka.Consumer
	topic       string
	deliveryCh  chan kafka.Event
	responseMap map[string]chan *eventGetResponse
	sync.Mutex
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

func NewOrderPlacer(conf *config2.ConfigKafka, groupID string) (*OrderPlacer, error) {
	configMapProducer := kafka.ConfigMap{
		"bootstrap.servers": conf.URL,
		"client.id":         "orderPlacer",
		"acks":              "all",
	}

	clientProducer, err := kafka.NewProducer(&configMapProducer)
	if err != nil {
		return nil, fmt.Errorf("kafka.NewProducer %w", err)
	}

	configMapConsumer := kafka.ConfigMap{
		"bootstrap.servers":  conf.URL,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	clientConsumer, err := kafka.NewConsumer(&configMapConsumer)
	if err != nil {
		clientProducer.Close()
		return nil, fmt.Errorf("ошибка при создании -> kafka.NewConsumer %w", err)
	}

	err = clientConsumer.Subscribe(conf.Topic+".event.response*", nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при подписке -> client.Subscribe %w", err)
	}

	return &OrderPlacer{
		producer:    clientProducer,
		consumer:    clientConsumer,
		topic:       conf.Topic,
		deliveryCh:  make(chan kafka.Event, 10000),
		responseMap: make(map[string]chan *eventGetResponse),
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

	if err != nil {
		return nil, fmt.Errorf("сообщение не было отправлено %w", err)
	}

	return op.awaitResponse(correlationID)
}

func (op *OrderPlacer) awaitResponse(correlationID string) (*entity2.Order, error) {
	responseCh := make(chan *eventGetResponse, 1)
	defer close(responseCh)

	op.Lock()
	op.responseMap[correlationID] = responseCh
	op.Unlock()

	select {
	case eventResponse := <-responseCh:
		if eventResponse == nil {
			return nil, fmt.Errorf("ответ по CorrelationID %s пустой", correlationID)
		}

		if correlationID != eventResponse.CorrelationID {
			return nil, fmt.Errorf("запрашевыемый сorrelationID и сorrelationID ответа не совпадают %s != %s",
				correlationID, eventResponse.CorrelationID)
		}

		if eventResponse.Status != true {
			return nil, fmt.Errorf("ошибка при поиске CorrelationID eventResponse.Status == false %s", correlationID)
		}

		return &eventResponse.Order, nil

	case <-time.After(time.Second * 10):
		op.Lock()
		delete(op.responseMap, correlationID)
		op.Unlock()
		return nil, fmt.Errorf("ответ по CorrelationID %s не получен вовремя", correlationID)
	}
}

func (op *OrderPlacer) handleResponse(response *eventGetResponse) {
	op.Lock()
	responseCh, ok := op.responseMap[response.CorrelationID]
	op.Unlock()

	if ok {
		responseCh <- response
		close(responseCh)

		op.Lock()
		delete(op.responseMap, response.CorrelationID)
		op.Unlock()
	}
}

func (op *OrderPlacer) ListenResponse(ctx context.Context) {
	commit := func(msg *kafka.Message) {
		if _, err := op.consumer.CommitMessage(msg); err != nil {
			log.Printf("Коммит не выполнен: %s", err)
		}
	}

	run := true

	for run {
		select {
		case <-ctx.Done():
			log.Printf("Обработчик ответов остановлен...")
			run = false
			break
		default:
			msg, ok := op.consumer.Poll(150).(*kafka.Message)
			if !ok {
				continue
			}

			var evt event

			if err := json.NewDecoder(bytes.NewReader(msg.Value)).Decode(&evt); err != nil {
				log.Printf("Ошибка при декодировании event: %s", err)

				commit(msg)

				continue
			}

			//ok = false

			switch evt.Type {
			case "orders.event.response":
				var responseEvent eventGetResponse
				if err := json.Unmarshal(evt.Value, &responseEvent); err != nil {
					log.Printf("Ошибка при декодировании createEvent: %s", err)
					commit(msg)
					continue
				}

				op.handleResponse(&responseEvent)
				commit(msg)

			default:
				log.Printf("Неизвестный тип события: %s", evt.Type)
				commit(msg)
			}

		}

	}
}

func (op *OrderPlacer) Close() error {
	op.producer.Close()
	return op.consumer.Close()
}
