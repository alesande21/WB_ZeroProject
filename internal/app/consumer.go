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
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunConsumer() {
	//Загрузка конфига
	log.Println("Загрузка конфига...")
	config, err := config2.GetDefaultConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Проблема с загрузкой конфига\n: %s", err)
		return
	}

	// Инициализация базы данных
	log.Println("Инициализация базы данных...")
	var conn *database2.DBConnection
	conn, err = database2.Open(config.GetDBsConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "проблемы с драйвером подключения на этапе открытия\n: %s", err)
		return
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Ошибка при закрытии соединения с базой данных: %s", err)
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
	log.Println("Инициализация репозитория...")
	var postgresRep database2.DBRepository
	postgresRep, err = database2.CreatePostgresRepository(conn.GetConn)
	if err != nil {
		log.Printf("проблемы с инициализацией PostgresRepository: %s\n", err)
		return
	}

	// Инициализация сервиса
	log.Println("Инициализация сервиса...")
	cache := database2.NewCache()
	orderRepo := repository2.NewOrderRepo(postgresRep, cache)
	orderService := service2.NewOrderService(orderRepo)

	// Обновление кеша
	go orderRepo.ListenForDbChanges(ctx, updateCache)

	log.Println("Загрузка конфига для подключения к кафке...")
	configKafka, err := kafka2.GetConfigProducer()
	if err != nil {
		log.Printf("Проблема с загрузкой конфига: %s", err.Error())
		return
	}

	log.Println("Подключение к кафке...")
	consumer, err := kafka2.NewOrderConsumer(configKafka, orderService, "order-service")
	if err != nil {
		log.Printf("Проблема с подключением к кафке: %s", err.Error())
		return
		//return nil, fmt.Errorf("internal.app.NewOrderConsumer %w", err)
	}
	defer consumer.Close()

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
