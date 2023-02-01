package oss

import (
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"io"
)

var bucket = initialization.GetBucket()

func UploadFromFile(ossPath, localFilePath string) error {
	return bucket.PutObjectFromFile(ossPath, localFilePath)
}

func UploadFromReader(ossPath string, srcReader io.Reader) error {
	return bucket.PutObject(ossPath, srcReader)
}
