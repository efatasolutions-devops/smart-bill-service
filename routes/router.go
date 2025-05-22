package routes

import (
	"github.com/arifin2018/splitbill-arifin.git/injector"
	"github.com/gofiber/fiber/v2"
)

func Router(app *fiber.App) {
	allController := injector.InitializeController()

	app.Post("/", allController.SplitbilController.Splitbil)
}
