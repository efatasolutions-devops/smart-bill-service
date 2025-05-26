package files

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"

	"github.com/gofiber/fiber/v2"
)

var ImageLocation = "./storage/public/images"

func UploadImage(app *fiber.Ctx, fileheader *multipart.FileHeader, imagePath string) error {
	// 1. Get the file from the form-data

	// 2. Open the uploaded file
	file, err := fileheader.Open()
	if err != nil {
		return app.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error opening file: %v", err))
	}
	defer file.Close()

	// 3. Read the file data
	imageData, err := ioutil.ReadAll(file)
	if err != nil {
		return app.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error reading file: %v", err))
	}

	// 4. Process the image data (e.g., save to disk)
	// filename := fmt.Sprintf("./storage/public/images/%s", fileheader.Filename) // Create a filename
	err = os.WriteFile(imagePath, imageData, 0666)
	if err != nil {
		return app.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error saving file: %v", err))
	}
	return nil
}
