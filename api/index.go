package handler

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/arifin2018/splitbill-arifin.git/config"
	appconfig "github.com/arifin2018/splitbill-arifin.git/config/appConfig"
	"github.com/arifin2018/splitbill-arifin.git/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Handler is the main entry point of the application. Think of it like the main() method
func Handler(w http.ResponseWriter, r *http.Request) {
	// This is needed to set the proper request path in `*fiber.Ctx`
	r.RequestURI = r.URL.String()

	handler().ServeHTTP(w, r)
}

// building the fiber application
func handler() http.HandlerFunc {
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

	config.Logger(app)
	routes.Router(app)
	if err := app.Listen(":3000"); err != nil {
		panic(err.Error())
	}

	<-serverShutdown

	config.GeneralLogger.Println("Running cleanup tasks...")
	fmt.Println("Running cleanup tasks...")

	return adaptor.FiberApp(app)
}
