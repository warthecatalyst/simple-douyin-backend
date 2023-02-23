package main

import (
	"fmt"
	"github.com/Shopify/sarama"
)

// kafka consumer

func main2() {
	consumer, err := sarama.NewConsumer([]string{"127.0.0.1:9092"}, nil)
	if err != nil {
		fmt.Printf("fail to start consumer, err:%v\n", err)
		return
	}
	partitionList, err := consumer.Partitions("test") // 根据topic取到所有的分区
	if err != nil {
		fmt.Printf("fail to get list of partition:err%v\n", err)
		return
	}
	fmt.Println(partitionList)
	for _, partition := range partitionList { // 遍历所有的分区
		fmt.Println("partition = ", partition)
		pc, err := consumer.ConsumePartition("test", partition, sarama.OffsetNewest)
		if err != nil {
			fmt.Printf("failed to start consumer for partition %d,err:%v\n", partition, err)
			return
		}
		defer pc.AsyncClose()
		// 异步从每个分区消费信息
		for msg := range pc.Messages() {
			msg1 := msg
			go func() {
				fmt.Println("Got messageFrom 消息队列")
				value := string(msg1.Value)
				fmt.Printf("Partition:%d Offset:%d Key:%v Value:%v\n", msg1.Partition, msg1.Offset, msg1.Key, value)
			}()
		}
	}
}

func main() {
	main2()
}
