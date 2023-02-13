package init

import (
	"fmt"
	"github.com/Shopify/sarama"
)

var kafkaServer sarama.SyncProducer
var kafkaClient sarama.Consumer

func InitKafkaServer() {
	var err error
	config := sarama.NewConfig()
	switch kafkaServerConf.RequireACKs {
	case "NoResponse":
		config.Producer.RequiredAcks = sarama.NoResponse
	case "WaitForLocal":
		config.Producer.RequiredAcks = sarama.WaitForLocal
	case "WaitForAll":
		config.Producer.RequiredAcks = sarama.WaitForAll
	}
	switch kafkaServerConf.Partitioner {
	case "NewRandomPartitioner":
		config.Producer.Partitioner = sarama.NewRandomPartitioner
	}
	config.Producer.Return.Successes = kafkaServerConf.ReturnSuccesses
	kafkaServer, err = sarama.NewSyncProducer([]string{fmt.Sprintf("%s:%s", kafkaServerConf.Host, kafkaServerConf.Port)}, config)
	if err != nil {
		stdOutLogger.Panic().Caller().Str("Error occurs in InitKafkaServer,", err.Error())
	}
}

func InitKafkaClient() {
	var err error
	kafkaClient, err = sarama.NewConsumer([]string{fmt.Sprintf("%s:%s", kafkaClientConf.Host, kafkaClientConf.Port)}, nil)
	if err != nil {
		stdOutLogger.Panic().Caller().Str("Error occurs in InitKafkaClient,", err.Error())
	}
}

func GetKafkaServer() sarama.SyncProducer {
	return kafkaServer
}

func GetKafkaClient() sarama.Consumer {
	return kafkaClient
}
