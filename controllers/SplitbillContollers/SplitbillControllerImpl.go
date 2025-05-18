package splitbillcontollers

import (
	"github.com/arifin2018/splitbill-arifin.git/helpers"
	"github.com/gofiber/fiber/v2"
)

func (splitbillControllerImpl *SplitbillControllerImpl) Splitbil(app *fiber.Ctx) error {
	jsonData, err := splitbillControllerImpl.SplitbillService.Splitbil(app)
	if err != nil {
		return helpers.ResultFailedJsonApi(app, jsonData, err.Error())
	}
	return helpers.ResultSuccessJsonApi(app, jsonData)
}
