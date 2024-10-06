package kafka

import (
	entity2 "WB_ZeroProject/internal/entity"
	"WB_ZeroProject/internal/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ilyakaznacheev/cleanenv"
	log2 "github.com/sirupsen/logrus"
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

type ConfigKafka struct {
	URL   string `env-required:"true" json:"url" yaml:"url" env:"KAFKA_CONN"`
	Host  string `env-required:"true" json:"host" yaml:"host" env:"KAFKA_HOST"`
	Port  int    `env-required:"true" json:"port" yaml:"port" env:"KAFKA_PORT"`
	Topic string `env-required:"true" json:"topic" yaml:"topic" env:"KAFKA_TOPIC"`
}

func GetConfigProducer() (*ConfigKafka, error) {
	var newConf ConfigKafka
	if err := cleanenv.ReadEnv(&newConf); err != nil {
		return nil, fmt.Errorf("-> cleanenv.ReadEnv: ошибка загрузки env параметров конфига для kafka: %w", err)
	}

	return &newConf, nil
}

func NewOrderPlacer(conf *ConfigKafka, groupID string) (*OrderPlacer, error) {
	configMapProducer := kafka.ConfigMap{
		"bootstrap.servers": conf.URL,
		"client.id":         "orderPlacer",
		"acks":              "all",
	}

	clientProducer, err := kafka.NewProducer(&configMapProducer)
	if err != nil {
		return nil, fmt.Errorf("-> kafka.NewProducer: ошибка при инициализации нового Producer: %w", err)
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
		return nil, fmt.Errorf("-> kafka.NewConsumer: ошибка при инициализации нового Consumer: %w", err)
	}

	err = clientConsumer.Subscribe("orders.event.response", nil)
	if err != nil {
		return nil, fmt.Errorf("-> clientConsumer.Subscribe: ошибка при подписке на orders.event.response: %w", err)
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

	tipicReq := op.topic + ".request"

	err := op.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &tipicReq,
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
		return fmt.Errorf("-> json.NewEncoder: ошибка кодирования сообщения eventCreate: %w", err)
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
		return fmt.Errorf("-> op.producer.Produce: ошибка при отправке сообщения: %w", err)
	}

	return nil
}

// TODO: понять как вернуть ордер из кафка. Использовать ли мапу?
func (op *OrderPlacer) GetOrder(ctx context.Context, msgType string, orderId entity2.OrderId) (*entity2.Order, error) {

	var b bytes.Buffer
	correlationID, err := utils.GenerateUUIDV7()
	if err != nil {
		return nil, fmt.Errorf("-> utils.GenerateUUIDV7: ошибка при генерации uuid: %w", err)
	}

	evt := eventGet{
		Type:          msgType,
		Value:         orderId,
		CorrelationID: correlationID,
	}

	err = json.NewEncoder(&b).Encode(evt)

	if err != nil {
		return nil, fmt.Errorf("-> json.NewEncoder: ошибка кодирования сообщения eventGet: %w", err)

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
		return nil, fmt.Errorf("-> op.producer.Produce: ошибка при отправке сообщения: %w", err)
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
			return nil, fmt.Errorf("-> op.awaitResponse: ответ на запрос по CorrelationID %s пустой", correlationID)
		}

		if correlationID != eventResponse.CorrelationID {
			return nil, fmt.Errorf("-> op.awaitResponse: запрашевыемый сorrelationID и сorrelationID ответа не совпадают %s != %s",
				correlationID, eventResponse.CorrelationID)
		}

		if eventResponse.Status != true {
			return nil, fmt.Errorf("-> op.awaitResponse: ошибка при поиске CorrelationID %s eventResponse.Status == false", correlationID)
		}

		return &eventResponse.Order, nil

	case <-time.After(time.Second * 10):
		op.Lock()
		delete(op.responseMap, correlationID)
		op.Unlock()
		return nil, fmt.Errorf("-> op.awaitResponse: ответ по CorrelationID %s не получен вовремя", correlationID)
	}
}

func (op *OrderPlacer) ListenResponse(ctx context.Context) {
	commit := func(msg *kafka.Message) {
		if _, err := op.consumer.CommitMessage(msg); err != nil {
			log2.Errorf("ListenResponse-> op.consumer.CommitMessage: коммит не выполнен: %s", err)
		}
	}

	run := true

	for run {
		select {
		case <-ctx.Done():
			log2.Info("Обработчик ответов остановлен...")
			run = false
			break

		default:
			msg, ok := op.consumer.Poll(150).(*kafka.Message)
			if !ok {
				continue
			}

			var evt event
			if err := json.NewDecoder(bytes.NewReader(msg.Value)).Decode(&evt); err != nil {
				log2.Errorf("ListenResponse-> json.NewDecoder: ошибка при декодировании event: %s", err)
				commit(msg)
				continue
			}

			//ok = false

			switch evt.Type {
			case "orders.event.response":
				var responseEvent eventGetResponse
				if err := json.NewDecoder(bytes.NewReader(msg.Value)).Decode(&responseEvent); err != nil {
					log2.Errorf("ListenResponse-> json.NewDecoder: ошибка при декодировании responseEvent: %s", err)
					commit(msg)
					continue
				}

				op.handleResponse(&responseEvent)
				commit(msg)

			default:
				log2.Errorf("ListenResponse: неизвестный тип события: %s", evt.Type)
				commit(msg)
			}

		}

	}
}

func (op *OrderPlacer) handleResponse(response *eventGetResponse) {
	op.Lock()
	responseCh, ok := op.responseMap[response.CorrelationID]
	op.Unlock()

	if ok {
		responseCh <- response
		time.Sleep(2 * time.Second)
		op.Lock()
		delete(op.responseMap, response.CorrelationID)
		op.Unlock()
	}
}

func (op *OrderPlacer) Close() error {
	op.producer.Close()
	return op.consumer.Close()
}
