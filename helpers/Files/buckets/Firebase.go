package buckets

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"

	"github.com/arifin2018/splitbill-arifin.git/config"
	"github.com/arifin2018/splitbill-arifin.git/helpers/files/buckets/models"
	"github.com/disintegration/imaging"
)

type Firebase struct {
}

func (firebase *Firebase) CheckFileSizeAndResizeFileIfNecessary(imageData []byte) (imageDataReader models.ImagerDataReader, err error) {
	config.GeneralLogger.Printf("[Firebase Upload Info] File size (%d bytes) exceeds 2MB. Resizing...\n", len(imageData))

	// Dekode gambar
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		config.GeneralLogger.Printf("[Firebase Upload Error] Error decoding image for resizing: %v\n", err)
		return models.ImagerDataReader{}, fmt.Errorf("error decoding image for resizing: %w", err)
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
		return models.ImagerDataReader{}, fmt.Errorf("error encoding resized image: %w", err)
	}

	reader := bytes.NewReader(buf.Bytes()) // Gunakan data gambar yang sudah di-resize
	imageData = buf.Bytes()                // Update imageData dengan data yang sudah di-resize
	config.GeneralLogger.Printf("[Firebase Upload Info] Resizing complete. New size: %d bytes\n", buf.Len())
	return models.ImagerDataReader{
		Reader:    reader,
		ImageData: imageData,
	}, nil
}

func (firebase *Firebase) CreateFileStorageAndPublish(objectName string, imageDataReader models.ReaderFileHeader) (string, error) {
	// 3. Tentukan nama file di Firebase Storage
	// objectName sekarang dapat mengakses timestamp dan safeFilename

	// 4. Buat context untuk operasi Firebase Storage
	ctx := context.Background()

	// 5. Dapatkan writer untuk objek di Firebase Storage
	wc := config.FirebaseStorageBucket.Object(objectName).NewWriter(ctx)

	wc.ContentType = imageDataReader.Fileheader.Header.Get("Content-Type")

	// 6. Salin data file dari reader (yang mungkin sudah di-resize) ke Firebase Storage
	if _, err := io.Copy(wc, imageDataReader.Reader); err != nil {
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
