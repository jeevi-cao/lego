package kafka

import (
	"fmt"
	"testing"
)

func TestNewKafkaConsumer(t *testing.T) {

	kafkaConsumer := Consumer{}
	consumerSetting := &ConsumerSetting{
		Hosts:         "10.103.17.53:9092",
		Topic:         "test",
		ReturnSuccess: true,
	}
	config := buildConsumerConfig(consumerSetting)
	kafkaConsumer.config = config
	kafkaConsumer.setting = consumerSetting
	kafkaConsumer.consumerMsg(consumerSetting.Topic, consumerMsgFunc)
}

func consumerMsgFunc(msg string) {
	fmt.Println("consumer msg :" + msg)
}
