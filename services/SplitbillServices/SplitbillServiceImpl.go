package splitbillservices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arifin2018/splitbill-arifin.git/config"
	files "github.com/arifin2018/splitbill-arifin.git/helpers/Files"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/genai"
)

func (splitbilSeviceImpl *SplibillServiceImpl) Splitbil(app *fiber.Ctx) (map[string]interface{}, error) {
	t := time.Now()
	timestamp := t.Format("20060102150405")

	fileheader, err := app.FormFile("image")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error retrieving file: %v", err.Error()))
	}

	nameFileImage := fmt.Sprintf("%v_%v", timestamp, strings.TrimSpace(fileheader.Filename))
	imagePath := fmt.Sprintf("%s/%s", files.ImageLocation, nameFileImage) // Ganti dengan path gambar struk Indomaret Anda
	files.UploadImage(app, fileheader, imagePath)
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create client: %v", err.Error()))
	}
	config.GeneralLogger.Println(app.Query("image"))

	config.GeneralLogger.Println(imagePath)
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to read image: %v", err.Error()))
	}

	prompt := `Tolong lakukan Optical Character Recognition (OCR) pada gambar struk ini dan ekstrak informasi belanja. Kembalikan hasilnya dalam format JSON dengan struktur berikut:

{
  "items": [
    {
      "name": "[Nama Barang 1]",
      "price": "[Harga per Unit 1]",
      "quantity": "[Kuantitas 1]",
      "total": "[Total Harga Item 1]"
    },
    {
      "name": "[Nama Barang 2]",
      "price": "[Harga per Unit 2]",
      "quantity": "[Kuantitas 2]",
      "total": "[Total Harga Item 2]"
    }
    // ... (dan seterusnya untuk semua item)
  ],
  "store_information": {
    "address": "[Alamat Toko]",
    "email": "[Email Toko]",
    "npwp": "[NPWP Toko]",
    "phone_number": "[Nomor Telepon Toko]",
    "store_name": "[Nama Toko]"
  },
  "totals": {
    "change": "[Uang Kembali]",
    "discount": "[Nilai Diskon]",
    "payment": "[Jumlah Pembayaran]",
    "service_charge": "[Biaya Layanan]",
    "subtotal": "[Subtotal]",
    "tax": {
      "amount": "[Nilai Pajak]",
      "dpp": "[Dasar Pengenaan Pajak]",
      "name": "[Nama Pajak]"
    },
    "total": "[Total Belanja]"
  },
  "transaction_information": {
    "date": "[Tanggal Transaksi]",
    "time": "[Waktu Transaksi]",
    "transaction_id": "[ID Transaksi]"
  }
}

Pastikan semua nilai diisi sesuai dengan informasi yang tertera pada struk. Jika suatu informasi tidak ditemukan, gunakan nilai null atau string kosong untuk field yang sesuai. Untuk nilai numerik (harga, kuantitas, total, dll.), kembalikan dalam format string seperti pada contoh.
`

	parts := []*genai.Part{
		genai.NewPartFromText(prompt),
		&genai.Part{
			InlineData: &genai.Blob{
				MIMEType: "image/jpeg",
				Data:     imgData,
			},
		},
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		contents,
		nil,
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to generate content: %v", err.Error()))
	}

	responseText := result.Text()
	config.GeneralLogger.Println("Raw response from Gemini:")
	config.GeneralLogger.Println(responseText)

	// Bersihkan string dari karakter di luar JSON valid
	cleanedJSON := strings.TrimSpace(responseText)
	if strings.HasPrefix(cleanedJSON, "```json") {
		cleanedJSON = cleanedJSON[len("```json"):]
	}
	if strings.HasSuffix(cleanedJSON, "```") {
		cleanedJSON = cleanedJSON[:len(cleanedJSON)-len("```")]
	}
	cleanedJSON = strings.TrimSpace(cleanedJSON)

	// Coba parse cleanedJSON sebagai JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(cleanedJSON), &jsonData)
	if err != nil {
		config.GeneralLogger.Println("\nFailed to unmarshal JSON after cleaning:")
		return nil, errors.New(fmt.Sprintf("Failed to unmarshal JSON after cleaning: %v", err.Error()))
		// Jika masih gagal, mungkin format JSON dari Gemini tidak sepenuhnya valid
	} else {
		config.GeneralLogger.Println("\nSuccessfully unmarshaled JSON after cleaning:")
		// Sekarang Anda dapat bekerja dengan jsonData sebagai map
		// Contoh mengakses beberapa field:
		if items, ok := jsonData["items"].([]interface{}); ok {
			config.GeneralLogger.Printf("Number of items: %d\n", len(items))
			if len(items) > 0 {
				if firstItem, ok := items[0].(map[string]interface{}); ok {
					config.GeneralLogger.Printf("First item name: %v\n", firstItem["name"])
					config.GeneralLogger.Printf("First item price: %v\n", firstItem["price"])
				}
			}
		}
		if storeInfo, ok := jsonData["store_information"].(map[string]interface{}); ok {
			config.GeneralLogger.Printf("Store Name: %v\n", storeInfo["store_name"])
			config.GeneralLogger.Printf("Store Address: %v\n", storeInfo["address"])
		}
		if totals, ok := jsonData["totals"].(map[string]interface{}); ok {
			config.GeneralLogger.Printf("Total belanja: %v\n", totals["total"])
			if taxInfo, ok := totals["tax"].(map[string]interface{}); ok {
				config.GeneralLogger.Printf("Tax Amount: %v\n", taxInfo["amount"])
				config.GeneralLogger.Printf("Tax Name: %v\n", taxInfo["name"])
			}
		}
		if transactionInfo, ok := jsonData["transaction_information"].(map[string]interface{}); ok {
			config.GeneralLogger.Printf("Transaction Date: %v\n", transactionInfo["date"])
			config.GeneralLogger.Printf("Transaction ID: %v\n", transactionInfo["transaction_id"])
		}
	}
	return jsonData, nil
}
