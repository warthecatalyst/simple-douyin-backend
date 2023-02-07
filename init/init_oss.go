package init

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	bucket *oss.Bucket
)

func InitOSS() {
	client, err := oss.New("https://"+OssConf.Url, OssConf.AccessKeyID, OssConf.AccessKeySecret)
	if err != nil {
		stdOutLogger.Panic().Caller().Str("OSS初始化client失败", err.Error())
	}
	bucket, err = client.Bucket(OssConf.Bucket)
	if err != nil {
		stdOutLogger.Panic().Caller().Str("OSS初始化bucket失败", err.Error())
	}
}

func GetBucket() *oss.Bucket {
	return bucket
}
