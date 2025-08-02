package buckets

import (
	"github.com/arifin2018/splitbill-arifin.git/helpers/files/buckets/models"
)

type BucketInterface interface {
	CheckFileSizeAndResizeFileIfNecessary(imageData []byte) (imageDataReader models.ImagerDataReader, err error)
	CreateFileStorageAndPublish(objectName string, imageDataReader models.ReaderFileHeader) (string, error)
}
