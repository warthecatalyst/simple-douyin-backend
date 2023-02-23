package cronUtils

import (
	"github.com/robfig/cron/v3"
)

//CronLab 分布式定时任务所使用的组件
var CronLab *cron.Cron

// InitCron 启动分布式定时任务
func InitCron() {
	CronLab = cron.New()
	CronLab.Start()
}
