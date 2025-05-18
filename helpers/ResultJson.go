package helpers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func ResultSuccessJsonApi(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusAccepted).JSON(data)
}

func ResultSuccessCreateJsonApi(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(data)
}

func ResultSuccessUpdateJsonApi(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(data)
}

func ResultSuccessDeleteJsonApi(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusNoContent).JSON(data)
}

func ResultFailedJsonApi(c *fiber.Ctx, data any, errorMessage string) error {
	return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
		"data":   data,
		"status": errorMessage,
	})
}

func RecoverPanicContext(c *fiber.Ctx) error {
	if r := recover(); r != nil {
		err := fmt.Sprintf("Error occured %s", r)
		return ResultFailedJsonApi(c, nil, err)
	}
	return nil
}
