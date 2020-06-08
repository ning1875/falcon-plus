package kafka

import (
	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/polymetric/g"
)

var KafkaAsyncProducer sarama.AsyncProducer

func InitKafkaConn() {
	log.Println("InitKafkaConn\n")
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(g.Config().Kafka.KafkaNodes, config)
	if err != nil {
		log.Errorf("InitKafkaConn_error:%+v", err)
		return
	}
	KafkaAsyncProducer = producer
}
