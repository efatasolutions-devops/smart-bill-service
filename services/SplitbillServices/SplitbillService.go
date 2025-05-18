package splitbillservices

import "github.com/gofiber/fiber/v2"

type SplibillService interface {
	Splitbil(app *fiber.Ctx) (map[string]interface{}, error)
}

type SplibillServiceImpl struct {
}

func NewSplitbillServiceImpl() *SplibillServiceImpl {
	return &SplibillServiceImpl{}
}
