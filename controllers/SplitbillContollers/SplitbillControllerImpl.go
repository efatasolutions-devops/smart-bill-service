package splitbillcontollers

import (
	"github.com/arifin2018/splitbill-arifin.git/helpers"
	"github.com/gofiber/fiber/v2"
)

// Splitbil processes receipt image and extracts splitbill information
// @Summary Extract splitbill information from receipt image
// @Description Upload a receipt image and extract detailed splitbill information including items, store details, totals, and transaction information using OCR and AI
// @Tags Splitbill
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Receipt image file (jpg, jpeg, png)"
// @Success 202 {object} models.SplitbillResponse "Successfully processed receipt"
// @Failure 406 {object} models.ErrorResponse "Failed to process receipt"
// @Router / [post]
func (splitbillControllerImpl *SplitbillControllerImpl) Splitbil(app *fiber.Ctx) error {
	jsonData, err := splitbillControllerImpl.SplitbillService.Splitbil(app)
	if err != nil {
		return helpers.ResultFailedJsonApi(app, jsonData, err.Error())
	}
	return helpers.ResultSuccessJsonApi(app, jsonData)
}
