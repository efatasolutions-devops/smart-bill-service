package buckets

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"os"

	"github.com/arifin2018/splitbill-arifin.git/config"
	"github.com/arifin2018/splitbill-arifin.git/helpers/files/buckets/models"
	"github.com/disintegration/imaging"
)

type VM struct {
}

// func (vm *VM) CheckFileSizeAndResizeFileIfNecessary(imageData []byte) (imageDataReader models.ImagerDataReader, err error) {
func (vm *VM) CheckFileSizeAndResizeFileIfNecessary(imageData []byte) (imageDataReader models.ImagerDataReader, err error) {
	config.GeneralLogger.Printf("[Firebase Upload Info] File size (%d bytes) exceeds limit. Resizing...\n", len(imageData))
	var reader *bytes.Reader
	// img, _, err := image.Decode(bytes.NewReader(imageData))
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		if err != nil {
			config.GeneralLogger.Printf("[Firebase Upload Error] Error decoding image for resizing: %v\n", err)
			return models.ImagerDataReader{}, errors.New(fmt.Sprintf("error decoding image for resizing: %s", err))
		}

		if img.Bounds().Dx() > 1000 {
			img = imaging.Resize(img, 1000, 0, imaging.Lanczos)
		}

		var buf bytes.Buffer
		err = imaging.Encode(&buf, img, imaging.JPEG)
		if err != nil {
			config.GeneralLogger.Printf("[Firebase Upload Error] Error encoding resized image: %v\n", err)
			return models.ImagerDataReader{}, errors.New(fmt.Sprintf("error encoding resized image: %w", err))
		}

		reader = bytes.NewReader(buf.Bytes())
		config.GeneralLogger.Printf("[Firebase Upload Info] Resizing complete. New size: %d bytes\n", buf.Len())
	}
	return models.ImagerDataReader{
		Reader:    reader,
		ImageData: nil,
	}, nil
}

// func (vm *VM) CreateFileStorageAndPublish(objectName string, imageDataReader models.ReaderFileHeader) (string, error) {
func (vm *VM) CreateFileStorageAndPublish(objectName string, imageDataReader models.ReaderFileHeader) (string, error) {
	// 5. Create storage directory if not exists
	storagePath := "storage/public"
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return "", errors.New(fmt.Sprintf("error creating storage directory: %s", err))
	}

	// Create the file
	filePath := fmt.Sprintf("%s/%s", storagePath, objectName)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error creating file: %s", err))
	}
	defer dst.Close()

	// Copy the file data
	if _, err = io.Copy(dst, imageDataReader.Reader); err != nil {
		return "", errors.New(fmt.Sprintf("error copying file: %w", err))
	}

	// Return the relative path to the file
	publicURL := fmt.Sprintf("/storage/images/%s", objectName)
	config.GeneralLogger.Printf("[Upload Info] Upload successful. Path: %s\n", publicURL)
	return publicURL, nil
}
