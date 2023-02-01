package init

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
	"gorm.io/gorm"
	"os"
	"strings"
)

const (
	configFilePath = "./configs/config.ini"
)

// StdOutLogger 初始化标准输出的Logger
var StdOutLogger = zerolog.New(os.Stdout)

type OssConfig struct {
	Url             string
	Bucket          string
	BucketDirectory string
	AccessKeyID     string
	AccessKeySecret string
}

type VideoConfig struct {
	SavePath      string
	AllowedExts   []string
	UploadMaxSize int64
}

type UserConfig struct {
	PasswordEncrpted bool
}

type LogConfig struct {
	LogFileWritten bool
	LogFilePath    string
}

// 解析配置文件
var (
	Port       string // 服务启动端口
	dbHost     string // 数据库服务器主机
	dbPort     string // 数据服务器端口
	dbUser     string // 数据库用户
	dbPassWord string // 数据库密码
	dbName     string // 数据库名
	dbLogLevel string // 数据库日志打印级别

	RdbHost string // redis主机
	RdbPort string // redis端口

	FeedListLength int

	OssConf OssConfig

	VideoConf VideoConfig

	UserConf UserConfig

	LogConf LogConfig
)

func InitConfig() {
	StdOutLogger.Print("in basic initialization")
	f, err := ini.Load(configFilePath)
	if err != nil {
		log.Panic().Caller().Err(errors.New("配置文件初始化失败"))
	}

	loadServer(f)
	loadDb(f)
	loadRdb(f)
	loadFeed(f)
	loadOss(f)
	loadVideo(f)
	loadUser(f)
	loadLog(f)
}

// loadServer 加载服务器配置
func loadServer(file *ini.File) {
	s := file.Section("server")
	Port = s.Key("Port").MustString("8888")
}

// loadDb 加载数据库相关配置
func loadDb(file *ini.File) {
	s := file.Section("database")
	dbName = s.Key("DbName").MustString("douyin")
	dbPort = s.Key("DbPort").MustString("3306")
	dbHost = s.Key("DbHost").MustString("127.0.0.1")
	dbUser = s.Key("DbUser").MustString("")
	dbPassWord = s.Key("DbPassWord").MustString("")
	dbLogLevel = s.Key("LogLevel").MustString("error")
}

func loadRdb(file *ini.File) {
	s := file.Section("redisControl")
	RdbHost = s.Key("Host").MustString("127.0.0.1")
	RdbPort = s.Key("Port").MustString("6379")
}

func loadFeed(file *ini.File) {
	s := file.Section("feed")
	FeedListLength = s.Key("ListLength").MustInt(30)
}

func loadOss(file *ini.File) {
	s := file.Section("oss")
	OssConf.Url = s.Key("Url").MustString("")
	OssConf.Bucket = s.Key("Bucket").MustString("")
	OssConf.BucketDirectory = s.Key("BucketDirectory").MustString("")
	OssConf.AccessKeyID = s.Key("AccessKeyID").MustString("")
	OssConf.AccessKeySecret = s.Key("AccessKeySecret").MustString("")
}

func loadVideo(file *ini.File) {
	s := file.Section("video")
	VideoConf.SavePath = s.Key("SavePath").MustString("../userdata/")
	videoExts := s.Key("AllowedExts").MustString("mp4,wmv,avi")
	VideoConf.AllowedExts = strings.Split(videoExts, ",")
	VideoConf.UploadMaxSize = s.Key("UploadMaxSize").MustInt64(1024)
}

func loadUser(file *ini.File) {
	s := file.Section("user")
	UserConf.PasswordEncrpted = s.Key("PasswordEncrypted").MustBool(false)
}

func loadLog(file *ini.File) {
	s := file.Section("log")
	LogConf.LogFileWritten = s.Key("FileLogWritten").MustBool(false)
	LogConf.LogFilePath = s.Key("LogFilePath").MustString("./logdata/logFile.txt")
}

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}
