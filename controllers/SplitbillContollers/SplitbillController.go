package splitbillcontollers

import (
	splitbillservices "github.com/arifin2018/splitbill-arifin.git/services/SplitbillServices"
	"github.com/gofiber/fiber/v2"
)

type SplitbilController interface {
	Splitbil(app *fiber.Ctx) error
}

type SplitbillControllerImpl struct {
	SplitbillService splitbillservices.SplibillService
}

func NewSplitbilController(splitbillService splitbillservices.SplibillService) *SplitbillControllerImpl {
	return &SplitbillControllerImpl{
		SplitbillService: splitbillService,
	}
}
