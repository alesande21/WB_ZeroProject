package app

import (
	"WB_ZeroProject/internal/colorAttribute"
	http2 "WB_ZeroProject/internal/controllers/http"
	kafka2 "WB_ZeroProject/internal/kafka"
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	log2 "github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunProducer() error {

	// Загрузка конфига для кафки
	log2.Info("Загрузка конфига для подключения к кафке...")
	config, err := kafka2.GetConfigProducer()
	if err != nil {
		return fmt.Errorf("-> config2.GetConfigProducer%w", err)
	}

	log2.Info("Подключение к кафке...")
	producer, err := kafka2.NewOrderPlacer(config, "order-placer")
	if err != nil {
		return fmt.Errorf("-> kafka2.NewOrderPlacer%w", err)
	}

	defer func(producer *kafka2.OrderPlacer) {
		err := producer.Close()
		if err != nil {

		}
	}(producer)

	// TODO: убрать или найти другое применение
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log2.Info("Активация оброботчика ответов...")
	go producer.ListenResponse(ctx)

	log2.Info("Загрузка настроек для сервера...")
	var serverAddress http2.ServerAddress
	//err = serverAddress.LoadConfigAddress("src/internal/controllers/http/config.yml")
	err = serverAddress.UpdateEnvAddress()
	if err != nil {
		fmt.Fprintf(os.Stderr, "настройки адреса сервера не загрузились\n: %s", err)
		return nil
	}

	log.Println("Инициализация и старт сервера...")
	swagger, err := http2.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка загрузки сваггера\n: %s", err)
		return nil
	}
	swagger.Servers = nil

	tenderServer := http2.NewTenderServer(producer)

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	//r.Use(middleware.OapiRequestValidator(swagger))
	http2.HandlerFromMux(tenderServer, r)

	s := &http.Server{
		Addr:    serverAddress.EnvAddress,
		Handler: r,
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	shutDownChan := make(chan error, 1)

	go func() {
		shutDownChan <- s.ListenAndServe()
	}()

	log.Printf("Подключнеие установлено -> %s", colorAttribute.ColorString(colorAttribute.FgYellow, serverAddress.EnvAddress))

	select {
	case sig := <-interrupt:
		log.Printf("Приложение прерывается: %s", sig)

		ctxShutDown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)

		// TODO: активировать в случает если контекст понадыбится
		//cancel()

		defer cancelShutdown()
		err := s.Shutdown(ctxShutDown)
		if err != nil {
			log.Printf("Ошибка при завершении сервера: %v", err)
		}

		log.Println("Сервер завершил работу")
	case err := <-shutDownChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка при запуске сервера: %s", err)
		}
	}

	return nil
}
