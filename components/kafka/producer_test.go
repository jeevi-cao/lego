package kafka

import (
	"fmt"
	"testing"
)

func TestNewKafkaProducer(t *testing.T) {
	producerSetting := &ProducerSetting{
		Hosts:         "10.103.17.53:9092",
		Topic:         "test",
		ReturnSuccess: true,
		RequiredAcks:  0,
	}
	config := buildProducerConfig(producerSetting)
	producer := &Producer{
		config:  config,
		setting: producerSetting,
	}
	partition, offset, error := producer.sendMsgSync("test", "111", "22222")
	if error != nil {
		fmt.Println(error.Error())
	}
	fmt.Sprintf("send msg success  partition：%d  offset：%d\n", partition, offset)
}
