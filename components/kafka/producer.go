package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"strings"
	"time"
)

type Producer struct {
	config  *sarama.Config
	setting *ProducerSetting
}

type ProducerSetting struct {
	Hosts         string
	Topic         string
	ReturnSuccess bool
	RequiredAcks  int
	Timeout       int
	HostArr       []string
}

func NewKafkaProducer(producerSetting *ProducerSetting) *Producer {
	config := buildProducerConfig(producerSetting)
	return &Producer{config, producerSetting}
}

func buildProducerConfig(producerSetting *ProducerSetting) *sarama.Config {
	hosts := producerSetting.Hosts
	if len(hosts) > 0 {
		producerSetting.HostArr = strings.Split(hosts, ",")
	}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = producerSetting.ReturnSuccess
	if producerSetting.Timeout > 0 {
		config.Producer.Timeout = time.Duration(producerSetting.Timeout) * time.Second
	}
	switch producerSetting.RequiredAcks {
	case -1:
		config.Producer.RequiredAcks = sarama.WaitForAll
	case 0:
		config.Producer.RequiredAcks = sarama.NoResponse
	case 1:
		config.Producer.RequiredAcks = sarama.WaitForLocal
	}
	return config
}

func (kafkaProducer *Producer) sendMsgSync(topic string, key string, value string) (partition int32, offset int64, err error) {
	syncProducer, err := sarama.NewSyncProducer(kafkaProducer.setting.HostArr, kafkaProducer.config)
	if err != nil {
		return int32(-1), int64(-1), err
	}
	defer syncProducer.Close()
	msg := sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(key), Value: sarama.ByteEncoder(value)}
	partition, offset, error := syncProducer.SendMessage(&msg)
	return partition, offset, error
}

func (kafkaProducer *Producer) sendMsgASync(topic string, key string, value string) error {
	asyncProducer, err := sarama.NewAsyncProducer(kafkaProducer.setting.HostArr, kafkaProducer.config)
	if err != nil {
		return err
	}
	defer asyncProducer.Close()
	msg := sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(key), Value: sarama.ByteEncoder(value)}
	asyncProducer.Input() <- &msg
	select {
	case suc := <-asyncProducer.Successes():
		fmt.Printf("offset: %d,  timestamp: %s", suc.Offset, suc.Timestamp.String())
		return nil
	case fail := <-asyncProducer.Errors():
		fmt.Printf("err: %s\n", fail.Err.Error())
		return fail
	}
}
