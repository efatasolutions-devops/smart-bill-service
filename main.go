package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/arifin2018/splitbill-arifin.git/config"
	appconfig "github.com/arifin2018/splitbill-arifin.git/config/appConfig"
	"github.com/arifin2018/splitbill-arifin.git/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	serverShutdown := make(chan struct{})

	go func() {
		_ = <-c
		fmt.Println("Gracefully shutting down...")
		_ = app.Shutdown()
		serverShutdown <- struct{}{}
	}()
	appconfig.InitApplication()
	app.Use(cors.New())

	config.ConnectFirebase()
	config.Logger(app)
	routes.Router(app)
	if err := app.Listen(":3000"); err != nil {
		panic(err.Error())
	}

	<-serverShutdown

	config.GeneralLogger.Println("Running cleanup tasks...")
	fmt.Println("Running cleanup tasks...")
}
