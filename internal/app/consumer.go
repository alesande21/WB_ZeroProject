package app

import (
	config2 "WB_ZeroProject/internal/config"
	database2 "WB_ZeroProject/internal/database"
	kafka2 "WB_ZeroProject/internal/kafka"
	repository2 "WB_ZeroProject/internal/repository"
	service2 "WB_ZeroProject/internal/service"
	"context"
	"errors"
	"fmt"
	log2 "github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunConsumer() error {

	// Настройка логера
	SetLevel("debug", "console")
	log2.Info("Настройка логера...")

	//Загрузка конфига
	log.Println("Загрузка конфига для базы данных...")
	config, err := config2.GetDefaultConfig()
	if err != nil {
		return fmt.Errorf("-> config2.GetDefaultConfig%w", err)
	}

	// Инициализация базы данных
	log.Println("Инициализация базы данных...")
	var conn *database2.DBConnection
	conn, err = database2.Open(config.GetDBsConfig())
	if err != nil {
		return fmt.Errorf("-> database2.Open%w", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log2.Infof("RunConsumer-> conn.Close:%s", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: убрать
	//go conn.InterapterConn()

	updateCache := make(chan interface{})
	defer close(updateCache)

	// Проверка подключения
	go conn.CheckConn(ctx, config.GetDBsConfig(), updateCache)

	// Инициализация репозитория
	log2.Info("Инициализация репозитория...")
	var postgresRep database2.DBRepository
	postgresRep, err = database2.CreatePostgresRepository(conn.GetConn)
	if err != nil {
		return fmt.Errorf("-> database2.CreatePostgresRepository%w", err)
	}

	// Инициализация сервиса
	log2.Info("Инициализация сервиса...")
	cache := database2.NewCache()
	orderRepo := repository2.NewOrderRepo(postgresRep, cache)
	orderService := service2.NewOrderService(orderRepo)

	// Обновление кеша
	go orderRepo.ListenForDbChanges(ctx, updateCache)

	log2.Info("Загрузка конфига для подключения к кафке...")
	configKafka, err := kafka2.GetConfigProducer()
	if err != nil {
		return fmt.Errorf("-> kafka2.GetConfigProducer%w", err)
	}

	log2.Info("Подключение к кафке...")
	consumer, err := kafka2.NewOrderConsumer(configKafka, orderService, "order-service")
	if err != nil {
		return fmt.Errorf("->  kafka2.NewOrderConsumer%w", err)
	}
	defer func(consumer *kafka2.OrderConsumer) {
		err := consumer.Close()
		if err != nil {
			log2.Errorf("RunConsumer-> consumer.Close: ошибка при закрытии Consumer: %s", err.Error())
		}
	}(consumer)

	go consumer.ListenAndServe(ctx)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	shutDownChan := make(chan error, 1)

	// TODO: некорректно завершается найти причину
	for {

		select {
		case sig := <-interrupt:
			log.Printf("Приложение прерывается: %s", sig)

			//ctxShutDown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
			cancel()
			time.Sleep(10 * time.Second)

			//defer cancelShutdown()
			//err := s.Shutdown(ctxShutDown)
			//if err != nil {
			//	log.Printf("Ошибка при завершении сервера: %v", err)
			//}

			log.Println("Обработчик завершил работу работу")
		case err := <-shutDownChan:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("Ошибка при запуске сервера: %s", err)
			}
		}
	}

}
