package logger

import (
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/rs/zerolog"
	"os"
)

// FileLogger 写入文件的Logger
var FileLogger zerolog.Logger

// InitFileLogger 初始化文件Logger
func InitFileLogger(config initialization.LogConfig) {
	var err error
	var file *os.File
	if config.LogFileWritten {
		if checkFileIsExist(config.LogFilePath) {
			file, err = os.OpenFile(config.LogFilePath, os.O_APPEND, 0666)
		} else {
			file, err = os.Create(config.LogFilePath)
		}
	}

	if err != nil {
		initialization.StdOutLogger.Panic().Caller().Err(err)
	}

	FileLogger = zerolog.New(file)
}

func checkFileIsExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}
