package app

import (
	"WB_ZeroProject/internal/colorAttribute"
	config2 "WB_ZeroProject/internal/config"
	http2 "WB_ZeroProject/internal/controllers/http"
	database2 "WB_ZeroProject/internal/database"
	repository2 "WB_ZeroProject/internal/repository"
	service2 "WB_ZeroProject/internal/service"
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	middleware "github.com/oapi-codegen/nethttp-middleware"
)

func InitSchema(repo database2.DBRepository, pathToSchema string) error {
	schema, err := os.ReadFile(pathToSchema)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = repo.Exec(ctx, string(schema))
	if err != nil {
		return err
	}

	return nil
}

func Run() {

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
	//defer func(conn *database2.DBConnection) {
	//	err := conn.Close()
	//	if err != nil {
	//
	//	}
	//}(conn)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Ошибка при закрытии соединения с базой данных: %s", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go conn.InterapterConn()

	updateCache := make(chan interface{})
	defer close(updateCache)
	go conn.CheckConn(ctx, config.GetDBsConfig(), updateCache)

	// Инициализация репозитория
	log.Println("Инициализация репозитория...")
	var postgresRep database2.DBRepository
	postgresRep, err = database2.CreatePostgresRepository(conn.GetConn)
	if err != nil {
		log.Printf("проблемы с инициализацией PostgresRepository: %s\n", err)
		return
	}

	//err = InitSchema(postgresRep, "migrations/schema.sql")
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "не удалось загрузить схему \n: %s", err)
	//	return
	//}

	log.Println("Инициализация сервиса...")
	cache := database2.NewCache()
	orderRepo := repository2.NewOrderRepo(postgresRep, cache)
	orderService := service2.NewOrderService(orderRepo)

	go orderRepo.ListenForDbChanges(ctx, updateCache)

	log.Println("Загрузка настроек для сервера...")
	var serverAddress http2.ServerAddress
	//err = serverAddress.LoadConfigAddress("src/internal/controllers/http/config.yml")
	err = serverAddress.UpdateEnvAddress()
	if err != nil {
		fmt.Fprintf(os.Stderr, "настройки адреса сервера не загрузились\n: %s", err)
		return
	}

	log.Println("Инициализация и старт сервера...")
	swagger, err := http2.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка загрузки сваггера\n: %s", err)
		return
	}
	swagger.Servers = nil

	tenderServer := http2.NewTenderServer(orderService)

	//r := mux.NewRouter().PathPrefix("/api").Subrouter().StrictSlash(true)
	r := mux.NewRouter()

	//r.PathPrefix("/").Handler(http.FileServer(http.Dir("./internal/ui/")))

	sc := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/static/").Handler(sc)

	r.Use(middleware.OapiRequestValidator(swagger))
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
		cancel()
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
}
