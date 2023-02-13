package service

import (
	"github.com/Shopify/sarama"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"sync"
)

var (
	kafkaServer sarama.SyncProducer
	kafkaOnce   sync.Once
)

func initKafka() {
	kafkaOnce.Do(func() {
		kafkaServer = initialization.GetKafkaServer()
	})
}
