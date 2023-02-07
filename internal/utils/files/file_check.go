package files

import (
	"bytes"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"path"
	"strings"
)

// CheckFileExt 检查文件扩展名
func CheckFileExt(fileName string) bool {
	//检查文件的扩展名
	ext := path.Ext(fileName)
	ext = string(bytes.ToLower([]byte(ext)))
	for _, legalExt := range initialization.VideoConf.AllowedExts {
		if legalExt == ext {
			return true
		}
	}
	return false
}

// CheckFileSize 检查文件大小
func CheckFileSize(fileSize int64) bool {
	return fileSize <= initialization.VideoConf.UploadMaxSize*constants.MB
}

// GetFileNameWithoutExt 得到没有扩展名的文件
func GetFileNameWithoutExt(fileName string) string {
	ext := path.Ext(fileName)
	return strings.TrimSuffix(fileName, ext)
}
