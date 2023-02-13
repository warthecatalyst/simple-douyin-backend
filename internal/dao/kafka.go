package dao

import (
	"github.com/Shopify/sarama"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"sync"
)

var (
	kafkaClient sarama.Consumer
	kafkaOnce   sync.Once
)

func initKafkaClient() {
	kafkaOnce.Do(func() {
		kafkaClient = initialization.GetKafkaClient()
		go func() {
			for {
				err := GetFavoriteDaoInstance().getFromMessageQueue()
				if err == nil {
					break
				}
			}
		}()
	})
}
