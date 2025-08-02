package files

import (
	"bytes" // Import package bytes
	"context"
	"fmt"
	"image"        // Import package image
	_ "image/jpeg" // Penting: import driver codec untuk JPEG
	_ "image/png"  // Penting: import driver codec untuk PNG
	"io"
	"mime/multipart"
	"os"
	"strings" // Pastikan ini diimpor
	"time"

	"github.com/arifin2018/splitbill-arifin.git/config"
	"github.com/disintegration/imaging" // Import package imaging
	"github.com/gofiber/fiber/v2"
)

// Ukuran maksimum file yang diizinkan sebelum di-resize (dalam byte)
const MaxFileSizeBeforeResize = 1 * 1024 * 1024 // 2 MB

type UploadFileImpl struct {
}

// UploadImage mengunggah file ke Google Firebase Storage
func (uploadfileimpl UploadFileImpl) UploadImage(app *fiber.Ctx, fileheader *multipart.FileHeader) (string, error) {
	// 1. Dapatkan file dari form-data
	file, err := fileheader.Open()
	if err != nil {
		config.GeneralLogger.Printf("[Firebase Upload Error] Error opening file from form: %v\n", err)
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close() // Pastikan file ditutup setelah digunakan

	// Baca seluruh data file ke dalam buffer untuk pemeriksaan ukuran dan manipulasi
	imageData, err := io.ReadAll(file)
	if err != nil {
		config.GeneralLogger.Printf("[Firebase Upload Error] Error reading file data: %v\n", err)
		return "", fmt.Errorf("error reading file data: %w", err)
	}

	// Reset file reader ke awal setelah dibaca
	// Ini perlu karena file sudah dibaca io.ReadAll, jadi cursornya di akhir.
	// Jika file tidak memiliki method Seek, ini akan error.
	// Lebih aman menggunakan bytes.NewReader(imageData) untuk kedua kalinya.
	// file.Seek(0, 0) // Baris ini bisa dihapus karena kita akan menggunakan bytes.NewReader(imageData) lagi

	// --- PERBAIKAN SCOPE VARIABEL DI SINI ---
	// Pindahkan deklarasi t, timestamp, dan safeFilename ke sini
	t := app.Context().Time()
	timestamp := t.Format("20060102150405")
	safeFilename := strings.ReplaceAll(fileheader.Filename, " ", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "/", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "\\", "_")
	// --- AKHIR PERBAIKAN SCOPE VARIABEL ---

	// Inisialisasi reader untuk data gambar yang akan diunggah
	var reader io.Reader = bytes.NewReader(imageData) // Defaultnya adalah data asli

	// 2. Periksa ukuran file dan resize jika perlu
	if len(imageData) > MaxFileSizeBeforeResize {
		config.GeneralLogger.Printf("[Firebase Upload Info] File size (%d bytes) exceeds 2MB. Resizing...\n", len(imageData))

		// Dekode gambar
		img, _, err := image.Decode(bytes.NewReader(imageData))
		if err != nil {
			config.GeneralLogger.Printf("[Firebase Upload Error] Error decoding image for resizing: %v\n", err)
			return "", fmt.Errorf("error decoding image for resizing: %w", err)
		}

		// Opsi 1: Resize berdasarkan lebar maksimum dan jaga rasio aspek
		newWidth := img.Bounds().Dx()
		if newWidth > 1000 { // Jika gambar terlalu lebar, resize ke 1000px
			img = imaging.Resize(img, 1000, 0, imaging.Lanczos) // 0 berarti jaga rasio aspek
			// newWidth = 1000 // Variabel ini tidak lagi digunakan setelah ini, bisa dihapus
		}

		var buf bytes.Buffer
		// Coba simpan sebagai JPEG dengan kualitas lebih rendah
		// Pastikan imaging.WithQuality tersedia di versi imaging Anda
		err = imaging.Encode(&buf, img, imaging.JPEG) // Kualitas 75%
		if err != nil {
			config.GeneralLogger.Printf("[Firebase Upload Error] Error encoding resized image: %v\n", err)
			return "", fmt.Errorf("error encoding resized image: %w", err)
		}

		reader = bytes.NewReader(buf.Bytes()) // Gunakan data gambar yang sudah di-resize
		imageData = buf.Bytes()               // Update imageData dengan data yang sudah di-resize
		config.GeneralLogger.Printf("[Firebase Upload Info] Resizing complete. New size: %d bytes\n", buf.Len())
	}

	// 3. Tentukan nama file di Firebase Storage
	// objectName sekarang dapat mengakses timestamp dan safeFilename
	objectName := fmt.Sprintf("images/%s_%s", timestamp, safeFilename)
	fmt.Println(objectName)

	// 4. Buat context untuk operasi Firebase Storage
	ctx := context.Background()

	// 5. Dapatkan writer untuk objek di Firebase Storage
	wc := config.FirebaseStorageBucket.Object(objectName).NewWriter(ctx)
	wc.ContentType = fileheader.Header.Get("Content-Type")

	// 6. Salin data file dari reader (yang mungkin sudah di-resize) ke Firebase Storage
	if _, err = io.Copy(wc, reader); err != nil {
		if closeErr := wc.Close(); closeErr != nil {
			config.GeneralLogger.Printf("[Firebase Upload Error] Error closing writer after copy error: %v\n", closeErr)
		}
		return "", fmt.Errorf("error uploading file to Firebase Storage: %w", err)
	}

	// Pastikan writer ditutup setelah selesai menyalin
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("error closing Firebase Storage writer: %w", err)
	}

	// 7. Dapatkan URL publik dari file yang diunggah
	attrs, err := config.FirebaseStorageBucket.Attrs(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting bucket attributes: %w", err)
	}
	bucketName := attrs.Name

	publicURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media", bucketName, objectName)
	config.GeneralLogger.Printf("[Firebase Upload Info] Upload successful. Public URL: %s\n", publicURL)
	return publicURL, nil
}

func (uploadfileimpl UploadFileImpl) VM(app *fiber.Ctx, fileheader *multipart.FileHeader) (string, error) {
	// 1. Dapatkan file dari form-data
	file, err := fileheader.Open()
	if err != nil {
		config.GeneralLogger.Printf("[Firebase Upload Error] Error opening file from form: %v\n", err)
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// 2. Baca data file
	imageData, err := io.ReadAll(file)
	if err != nil {
		config.GeneralLogger.Printf("[Firebase Upload Error] Error reading file data: %v\n", err)
		return "", fmt.Errorf("error reading file data: %w", err)
	}

	// 3. Buat nama file dengan timestamp
	t := time.Now()
	timestamp := t.Format("20060102150405")
	safeFilename := strings.ReplaceAll(fileheader.Filename, " ", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "/", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "\\", "_")
	objectName := fmt.Sprintf("images/%s_%s", timestamp, safeFilename)

	// 4. Proses resize jika file terlalu besar
	var reader io.Reader = bytes.NewReader(imageData)
	if len(imageData) > MaxFileSizeBeforeResize {
		config.GeneralLogger.Printf("[Firebase Upload Info] File size (%d bytes) exceeds limit. Resizing...\n", len(imageData))

		img, _, err := image.Decode(bytes.NewReader(imageData))
		if err != nil {
			config.GeneralLogger.Printf("[Firebase Upload Error] Error decoding image for resizing: %v\n", err)
			return "", fmt.Errorf("error decoding image for resizing: %w", err)
		}

		if img.Bounds().Dx() > 1000 {
			img = imaging.Resize(img, 1000, 0, imaging.Lanczos)
		}

		var buf bytes.Buffer
		err = imaging.Encode(&buf, img, imaging.JPEG)
		if err != nil {
			config.GeneralLogger.Printf("[Firebase Upload Error] Error encoding resized image: %v\n", err)
			return "", fmt.Errorf("error encoding resized image: %w", err)
		}

		reader = bytes.NewReader(buf.Bytes())
		config.GeneralLogger.Printf("[Firebase Upload Info] Resizing complete. New size: %d bytes\n", buf.Len())
	}

	// 5. Create storage directory if not exists
	storagePath := "storage/public"
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return "", fmt.Errorf("error creating storage directory: %w", err)
	}

	// Create the file
	filePath := fmt.Sprintf("%s/%s", storagePath, objectName)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("error creating file: %w", err)
	}
	defer dst.Close()

	// Copy the file data
	if _, err = io.Copy(dst, reader); err != nil {
		return "", fmt.Errorf("error copying file: %w", err)
	}

	// Return the relative path to the file
	publicURL := fmt.Sprintf("/storage/images/%s", objectName)
	config.GeneralLogger.Printf("[Upload Info] Upload successful. Path: %s\n", publicURL)
	return publicURL, nil
}
