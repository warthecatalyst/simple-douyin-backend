package logger

import (
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/rs/zerolog/log"
	"testing"
)

func TestZLog(t *testing.T) {

	log.Debug().Caller().
		Str("Scale", "833 cents").
		Float64("Interval", 833.09).
		Msg("Fibonacci is everywhere")

	log.Debug().
		Str("Name", "Tom").
		Send()
}

func TestWriteToFileLog(t *testing.T) {
	logConf := initialization.LogConfig{
		LogFileWritten: true,
		LogFilePath:    "testFile.txt",
	}
	InitLogger(logConf)
	GlobalLogger.Print("Hello world")
}
