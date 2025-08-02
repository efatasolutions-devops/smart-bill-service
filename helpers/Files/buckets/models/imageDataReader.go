package models

import (
	"bytes"
	"io"
	"mime/multipart"
)

type ImagerDataReader struct {
	Reader    *bytes.Reader
	ImageData []byte
}

type ReaderFileHeader struct {
	Reader     io.Reader
	Fileheader *multipart.FileHeader
}
