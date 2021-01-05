package kafka

import (
	"github.com/Shopify/sarama"
	"strings"
	"sync"
)

var wg sync.WaitGroup

type Consumer struct {
	config  *sarama.Config
	setting *ConsumerSetting
}
type ConsumerSetting struct {
	Hosts         string
	Topic         string
	ReturnSuccess bool
	HostArr       []string
}

func NewKafkaConsumer(consumerSetting *ConsumerSetting) *Consumer {
	config := buildConsumerConfig(consumerSetting)
	return &Consumer{config, consumerSetting}
}

func buildConsumerConfig(consumerSetting *ConsumerSetting) *sarama.Config {
	hosts := consumerSetting.Hosts
	if len(hosts) > 0 {
		consumerSetting.HostArr = strings.Split(hosts, ",")
	}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = consumerSetting.ReturnSuccess
	return config
}

func (kafkaProducer *Consumer) consumerMsg(topic string, f func(msg string)) error {
	consumer, err := sarama.NewConsumer(kafkaProducer.setting.HostArr, kafkaProducer.config)
	if err != nil {
		return err
	}
	partitions, err := consumer.Partitions(topic)
	if err != nil {
		return err
	}
	for _, p := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(topic, p, sarama.OffsetNewest)
		if err != nil {
			continue
		}
		wg.Add(1)
		go func() {
			for msg := range partitionConsumer.Messages() {
				f(string(msg.Value))
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
}
