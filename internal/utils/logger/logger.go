package logger

import (
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/files"
	"github.com/rs/zerolog"
	"os"
)

// GlobalLogger 全局Logger
var GlobalLogger zerolog.Logger

// InitLogger 初始化GlobalLogger
func InitLogger(config initialization.LogConfig) {
	var err error
	var file *os.File
	if config.LogFileWritten {
		if exists, _ := files.PathExists(config.LogFilePath); exists {
			file, err = os.OpenFile(config.LogFilePath, os.O_APPEND, 0666)
		} else {
			file, err = os.Create(config.LogFilePath)
		}
		if err != nil {
			GlobalLogger = initialization.GetStdOutLogger()
			GlobalLogger.Error().Msg("Get Logger failed")
		}
		GlobalLogger = zerolog.New(file)
	} else {
		GlobalLogger = initialization.GetStdOutLogger()
	}
	GlobalLogger.Level(zerolog.InfoLevel)
}
