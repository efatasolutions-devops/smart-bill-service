package helpers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

func HandleConvertJSONMap(c *fiber.Ctx, jsonText string) error {
	// jsonText := "`json\n{\n  \"daftar_belanja\": [\n    {\n      \"no\": 1,\n      \"nama_barang\": \"GRNIER M.COOL FOAM50\",\n      \"harga_per_unit\": \"19900\",\n      \"kuantitas\": \"2\"\n    },\n    {\n      \"no\": 2,\n      \"nama_barang\": \"PLASTIK KCL\",\n      \"harga_per_unit\": \"1\",\n      \"kuantitas\": \"1\"\n    }\n  ],\n  \"total_belanja\": \"39,800\",\n  \"belanjaan_terdeteksi\": [\n    \"GRNIER M.COOL FOAM50\",\n    \"PLASTIK KCL\"\n  ]\n}\n`"

	cleanedJSON := jsonText[6 : len(jsonText)-1]

	var data map[string]interface{}
	err := json.Unmarshal([]byte(cleanedJSON), &data)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Gagal mengurai JSON",
		})
	}

	fmt.Printf("Data Daftar Belanja (map): %+v\n", data)
	return c.JSON(data)
}
