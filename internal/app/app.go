package app

import (
	"WB_ZeroProject/internal/colorAttribute"
	config2 "WB_ZeroProject/internal/config"
	http2 "WB_ZeroProject/internal/controllers/http"
	database2 "WB_ZeroProject/internal/database"
	repository2 "WB_ZeroProject/internal/repository"
	service2 "WB_ZeroProject/internal/service"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"

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

	go conn.InterapterConn()

	chanDbConn := make(chan *sql.DB)
	defer close(chanDbConn)
	go conn.CheckConn(config.GetDBsConfig(), chanDbConn)

	// Инициализация репозитория
	log.Println("Инициализация репозитория...")
	var postgresRep database2.DBRepository
	postgresRep, err = database2.CreatePostgresRepository(conn.GetConn2(), chanDbConn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "проблемы с созданием PostgresRepository \n: %s", err)
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

	r.Use(middleware.OapiRequestValidator(swagger))
	http2.HandlerFromMux(tenderServer, r)

	s := &http.Server{
		Addr:    serverAddress.EnvAddress,
		Handler: r,
	}

	log.Printf("Подключнеие установлено -> %s", colorAttribute.ColorString(colorAttribute.FgYellow, serverAddress.EnvAddress))
	log.Fatal(s.ListenAndServe())

}
