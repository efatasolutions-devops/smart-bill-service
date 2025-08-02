package files

import (
	"bytes"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/arifin2018/splitbill-arifin.git/config"
	"github.com/arifin2018/splitbill-arifin.git/helpers/files/buckets"
	"github.com/arifin2018/splitbill-arifin.git/helpers/files/buckets/models"
	"github.com/gofiber/fiber/v2"
)

// Ukuran maksimum file yang diizinkan sebelum di-resize (dalam byte)
const MaxFileSizeBeforeResize = 1 * 1024 * 1024 // 2 MB

type UploadFileImpl struct {
}

// UploadImage mengunggah file ke Google Firebase Storage
func (uploadfileimpl UploadFileImpl) UploadImage(app *fiber.Ctx, fileheader *multipart.FileHeader, bucket buckets.BucketInterface) (string, error) {
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
	t := time.Now()
	timestamp := t.Format("20060102150405")
	safeFilename := strings.ReplaceAll(fileheader.Filename, " ", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "/", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "\\", "_")
	// --- AKHIR PERBAIKAN SCOPE VARIABEL ---

	// Inisialisasi reader untuk data gambar yang akan diunggah
	var reader io.Reader = bytes.NewReader(imageData) // Defaultnya adalah data asli
	objectName := fmt.Sprintf("images/%s_%s", timestamp, safeFilename)

	// 2. Periksa ukuran file dan resize jika perlu
	if len(imageData) > MaxFileSizeBeforeResize {
		imageDataReader, err := bucket.CheckFileSizeAndResizeFileIfNecessary(imageData)
		if err != nil {
			return "", err
		}
		reader = imageDataReader.Reader

	}

	readerFileHeader := models.ReaderFileHeader{
		Reader:     reader,
		Fileheader: fileheader,
	}
	publicURL, err := bucket.CreateFileStorageAndPublish(objectName, readerFileHeader)
	if err != nil {
		return "", err
	}
	return publicURL, nil
}
