package files

import (
	"context"
	"fmt"
	"io"
	"log" // Tambahkan import log
	"mime/multipart"
	"strings"

	"github.com/arifin2018/splitbill-arifin.git/config"
	"github.com/gofiber/fiber/v2"
)

// UploadImage mengunggah file ke Google Firebase Storage
func UploadImage(app *fiber.Ctx, fileheader *multipart.FileHeader) (string, error) {
	// 1. Dapatkan file dari form-data
	file, err := fileheader.Open()
	if err != nil {
		log.Printf("[Firebase Upload Error] Error opening file from form: %v\n", err) // Log error
		return "", fmt.Errorf("error opening file: %w", err)                          // Kembalikan error yang lebih jelas
	}
	defer file.Close()

	// 2. Tentukan nama file di Firebase Storage
	t := app.Context().Time()
	timestamp := t.Format("20060102150405")
	safeFilename := strings.ReplaceAll(fileheader.Filename, " ", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "/", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "\\", "_")

	objectName := fmt.Sprintf("images/%s_%s", timestamp, safeFilename)

	// 3. Buat context untuk operasi Firebase Storage
	ctx := context.Background()

	// 4. Dapatkan writer untuk objek di Firebase Storage
	wc := config.FirebaseStorageBucket.Object(objectName).NewWriter(ctx)
	wc.ContentType = fileheader.Header.Get("Content-Type")

	// 5. Salin data file dari request ke Firebase Storage
	if _, err = io.Copy(wc, file); err != nil {
		// Pastikan writer ditutup meskipun ada error
		if closeErr := wc.Close(); closeErr != nil {
			log.Printf("[Firebase Upload Error] Error closing writer after copy error: %v\n", closeErr)
		}
		log.Printf("[Firebase Upload Error] Error copying file data to Firebase Storage: %v\n", err) // Log error
		return "", fmt.Errorf("error uploading file to Firebase Storage: %w", err)
	}

	// Pastikan writer ditutup setelah selesai menyalin
	if err := wc.Close(); err != nil {
		log.Printf("[Firebase Upload Error] Error closing Firebase Storage writer: %v\n", err) // Log error
		return "", fmt.Errorf("error closing Firebase Storage writer: %w", err)
	}

	// 6. Dapatkan URL publik dari file yang diunggah
	attrs, err := config.FirebaseStorageBucket.Attrs(ctx)
	if err != nil {
		log.Printf("[Firebase Upload Error] Error getting bucket attributes: %v\n", err) // Log error
		return "", fmt.Errorf("error getting bucket attributes: %w", err)
	}
	bucketName := attrs.Name

	publicURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media", bucketName, objectName)
	log.Printf("[Firebase Upload Info] Upload successful. Public URL: %s\n", publicURL) // Log sukses
	return publicURL, nil
}
