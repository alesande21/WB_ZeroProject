package server

import (
	"WB_ZeroProject/internal/colorAttribute"
	config2 "WB_ZeroProject/internal/config"
	"WB_ZeroProject/internal/database"
	"WB_ZeroProject/pkg/server/api"
	"flag"
	"fmt"
	middleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	Ip   = flag.String("ip", api.Localhost, "Set ip address")
	Port = flag.Int("port", api.DefaultPort, "Set instance port")
)

func main() {
	flag.Parse()

	swagger, err := api.GetSwagger()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	config, err2 := config2.GetDefaultConfig()
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "Problem with config\n: %s", err)
		return
	}

	conn, err := database.Open(config.GetDBsConfig())
	defer func(conn *database.DBConnection) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	rep, err := database.CreateRepository(conn.GetConn())

	if err != nil {
		fmt.Fprintf(os.Stderr, "Problem with CreateRepository\n: %s", err)
		return
	}

	ordersServer := api.OrdersServer{DB: rep}

	r := mux.NewRouter()

	r.Use(middleware.OapiRequestValidator(swagger))
	api.HandlerFromMux(&ordersServer, r)

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", *Ip, *Port),
		Handler: r,
	}

	log.Printf("Подключнеие установлено -> %s:%s", colorAttribute.ColorString(colorAttribute.FgYellow, *Ip),
		colorAttribute.ColorString(colorAttribute.FgYellow, strconv.Itoa(*Port)))
	log.Fatal(s.ListenAndServe())
}
