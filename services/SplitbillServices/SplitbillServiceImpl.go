package splitbillservices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil" // Tambahkan ini
	"os"
	"strings"

	// "time" // Tidak perlu lagi timestamp di sini, karena sudah di handle di UploadFile

	"github.com/arifin2018/splitbill-arifin.git/config"
	files "github.com/arifin2018/splitbill-arifin.git/helpers/Files"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/genai"
)

func (splitbilSeviceImpl *SplibillServiceImpl) Splitbil(app *fiber.Ctx) (map[string]interface{}, error) {
	fileheader, err := app.FormFile("image")
	if err != nil {
		config.GeneralLogger.Printf("Error retrieving file from form: %v\n", err.Error()) // Log lebih spesifik
		return nil, errors.New(fmt.Sprintf("Error retrieving file: %v", err.Error()))
	}

	uploadedImageURL, err := files.UploadImage(app, fileheader)
	if err != nil {
		// Ini akan mencetak error yang dikembalikan oleh files.UploadImage
		config.GeneralLogger.Printf("Failed to upload image to Firebase Storage: %v\n", err.Error())
		return nil, errors.New(fmt.Sprintf("Error uploading image to Firebase Storage: %v", err.Error()))
	}

	config.GeneralLogger.Println("Uploaded Image URL:", uploadedImageURL) // Ini harusnya tidak kosong jika tidak ada error

	// --- Perubahan besar di sini: Cara mendapatkan data gambar untuk Gemini ---
	file, err := fileheader.Open()
	if err != nil {
		config.GeneralLogger.Printf("Error opening file for Gemini: %v\n", err.Error()) // Log lebih spesifik
		return nil, errors.New(fmt.Sprintf("Error opening file for Gemini: %v", err.Error()))
	}
	defer file.Close()

	imgData, err := ioutil.ReadAll(file)
	if err != nil {
		config.GeneralLogger.Printf("Failed to read image data for Gemini: %v\n", err.Error()) // Log lebih spesifik
		return nil, errors.New(fmt.Sprintf("Failed to read image data for Gemini: %v", err.Error()))
	}
	// --- Akhir perubahan besar untuk Gemini ---

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		config.GeneralLogger.Printf("Failed to create Gemini client: %v\n", err.Error()) // Log lebih spesifik
		return nil, errors.New(fmt.Sprintf("Failed to create client: %v", err.Error()))
	}
	// config.GeneralLogger.Println(app.Query("image")) // Ini tidak relevan lagi

	config.GeneralLogger.Println("Uploaded Image URL:", uploadedImageURL) // Log URL gambar yang diunggah
	// config.GeneralLogger.Println(imagePath) // Ini tidak relevan lagi

	// Bagian prompt untuk Gemini tetap sama
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
    "discount": "[Nilai Diskon/Nilai Yang Dikurangi] kembalikan angka desimal tanpa pengurangan",
    "payment": "[Jumlah Pembayaran]",
    "subtotal": "[Subtotal]",
    "tax": {
      "amount": "[Nilai Pajak]",
      "service_charge": "[Biaya Layanan]",
      "dpp": "[Dasar Pengenaan Pajak]",
      "name": "[Nama Pajak]",
      "total_tax": "[Total Pajak dari service_charge + amount]"
    },
    "total": "[Total Belanja]"
  },
  "transaction_information": {
    "date": "[Tanggal Transaksi]",
    "time": "[Waktu Transaksi]",
    "transaction_id": "[ID Transaksi]"
  }
}

Pastikan semua nilai diisi sesuai dengan informasi yang tertera pada struk. Jika suatu informasi tidak ditemukan, gunakan nilai null atau string kosong untuk field yang sesuai. Untuk nilai numerik (harga, kuantitas, total, totals, discount, dll.), kembalikan dalam format desimal seperti pada contoh.
`

	parts := []*genai.Part{
		genai.NewPartFromText(prompt),
		&genai.Part{
			InlineData: &genai.Blob{
				MIMEType: fileheader.Header.Get("Content-Type"), // Gunakan Content-Type asli dari file header
				Data:     imgData,                               // Menggunakan imgData yang dibaca dari fileheader
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
	} else {
		config.GeneralLogger.Println("\nSuccessfully unmarshaled JSON after cleaning:")
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
