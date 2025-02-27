package init

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
	"gorm.io/gorm"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	configFilePath = "./configs/config.ini"
)

// stdOutLogger 初始化标准输出的Logger
var stdOutLogger = zerolog.New(os.Stdout)

type kafkaProducerConfig struct {
	Host            string
	Port            string
	RequireACKs     string
	Partitioner     string
	ReturnSuccesses bool
}

type KafkaConsumerConfig struct {
	Host string
	Port string
}

type ossConfig struct {
	Url             string
	Bucket          string
	BucketDirectory string
	AccessKeyID     string
	AccessKeySecret string
}

type videoConfig struct {
	SavePath      string
	AllowedExts   []string
	UploadMaxSize int64
}

type userConfig struct {
	PasswordEncrypted bool
}

type LogConfig struct {
	LogFileWritten bool
	LogFilePath    string
}

type RpcConfig struct {
	UserServiceHost     string
	UserServicePort     string
	VideoServiceHost    string
	VideoServicePort    string
	FavoriteServiceHost string
	FavoriteServicePort string
	CommentServiceHost  string
	CommentServicePort  string
	FollowServiceHost   string
	FollowServicePort   string
	MessageServiceHost  string
	MessageServicePort  string
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

	rdbHost string // redis主机
	rdbPort string // redis端口

	FeedListLength int

	kafkaServerConf kafkaProducerConfig
	kafkaClientConf KafkaConsumerConfig

	OssConf ossConfig

	VideoConf videoConfig

	UserConf userConfig

	LogConf LogConfig

	RpcCSConf RpcConfig
	RpcSDConf RpcConfig
)

func InitConfig() {
	stdOutLogger.Info().Msg("In initConfig")
	f, err := ini.Load(configFilePath)
	if err != nil {
		log.Panic().Caller().Err(errors.New("配置文件初始化失败"))
	}
	rand.Seed(time.Now().Unix())

	loadServer(f)
	loadDb(f)
	loadRdb(f)
	loadKafkaServer(f)
	loadKafkaClient(f)
	loadFeed(f)
	loadOss(f)
	loadVideo(f)
	loadUser(f)
	loadLog(f)
	loadRpcCSConf(f)
	loadRpcSDConf(f)
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
	s := file.Section("redis")
	rdbHost = s.Key("UserServiceHost").MustString("127.0.0.1")
	rdbPort = s.Key("Port").MustString("6379")
}

func loadKafkaServer(file *ini.File) {
	s := file.Section("kafkaProducer")
	kafkaServerConf.Host = s.Key("UserServiceHost").MustString("127.0.0.1")
	kafkaServerConf.Port = s.Key("Port").MustString("9092")
	kafkaServerConf.RequireACKs = s.Key("RequireACKs").MustString("WaitForAll")
	kafkaServerConf.Partitioner = s.Key("ProducerPartitioner").MustString("NewRandomPartitioner")
	kafkaServerConf.ReturnSuccesses = s.Key("ProducerReturnSuccesses").MustBool(true)
}

func loadKafkaClient(file *ini.File) {
	s := file.Section("kafkaConsumer")
	kafkaClientConf.Host = s.Key("UserServiceHost").MustString("127.0.0.1")
	kafkaClientConf.Port = s.Key("Port").MustString("9092")
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
	UserConf.PasswordEncrypted = s.Key("PasswordEncrypted").MustBool(false)
}

func loadLog(file *ini.File) {
	s := file.Section("log")
	LogConf.LogFileWritten = s.Key("FileLogWritten").MustBool(false)
	LogConf.LogFilePath = s.Key("LogFilePath").MustString("./logdata/logFile.txt")
}

func loadRpcCSConf(file *ini.File) {
	s := file.Section("rpcCS")
	RpcCSConf.UserServiceHost = s.Key("UserServiceHost").MustString("127.0.0.1")
	RpcCSConf.UserServicePort = s.Key("UserServicePort").MustString(":50051")
	RpcCSConf.VideoServiceHost = s.Key("VideoServiceHost").MustString("127.0.0.1")
	RpcCSConf.VideoServicePort = s.Key("VideoServicePort").MustString(":50052")
	RpcCSConf.FavoriteServiceHost = s.Key("FavoriteServiceHost").MustString("127.0.0.1")
	RpcCSConf.FavoriteServicePort = s.Key("FavoriteServicePort").MustString(":50053")
	RpcCSConf.CommentServiceHost = s.Key("CommentServiceHost").MustString("127.0.0.1")
	RpcCSConf.CommentServicePort = s.Key("CommentServicePort").MustString(":50054")
	RpcCSConf.FollowServiceHost = s.Key("FollowServiceHost").MustString("127.0.0.1")
	RpcCSConf.FollowServicePort = s.Key("FollowServicePort").MustString(":50055")
	RpcCSConf.MessageServiceHost = s.Key("MessageServiceHost").MustString("127.0.0.1")
	RpcCSConf.MessageServicePort = s.Key("MessageServicePort").MustString(":50056")

}

func loadRpcSDConf(file *ini.File) {
	s := file.Section("rpcSD")
	RpcSDConf.UserServiceHost = s.Key("UserServiceHost").MustString("127.0.0.1")
	RpcSDConf.UserServicePort = s.Key("UserServicePort").MustString(":50061")
	RpcSDConf.VideoServiceHost = s.Key("VideoServiceHost").MustString("127.0.0.1")
	RpcSDConf.VideoServicePort = s.Key("VideoServicePort").MustString(":50062")
	RpcSDConf.FavoriteServiceHost = s.Key("FavoriteServiceHost").MustString("127.0.0.1")
	RpcSDConf.FavoriteServicePort = s.Key("FavoriteServicePort").MustString(":50063")
	RpcSDConf.CommentServiceHost = s.Key("CommentServiceHost").MustString("127.0.0.1")
	RpcSDConf.CommentServicePort = s.Key("CommentServicePort").MustString(":50064")
	RpcSDConf.FollowServiceHost = s.Key("FollowServiceHost").MustString("127.0.0.1")
	RpcSDConf.FollowServicePort = s.Key("FollowServicePort").MustString(":50065")
	RpcSDConf.MessageServiceHost = s.Key("MessageServiceHost").MustString("127.0.0.1")
	RpcSDConf.MessageServicePort = s.Key("MessageServicePort").MustString(":50066")
}

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

func GetStdOutLogger() zerolog.Logger {
	return stdOutLogger
}
