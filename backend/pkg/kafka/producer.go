package kafka

import (
	"strings"

	"github.com/IBM/sarama"
)

func NewSyncProducer(cfg Config) (sarama.SyncProducer, error) {
	defaultCfg := sarama.NewConfig()

	defaultCfg.Producer.RequiredAcks = sarama.WaitForAll
	defaultCfg.Producer.Return.Successes = true

	return sarama.NewSyncProducer(strings.Split(cfg.Addrs, ","), defaultCfg)
}

func NewAsyncProducer(cfg Config) (sarama.AsyncProducer, error) {
	defaultCfg := sarama.NewConfig()
	return sarama.NewAsyncProducer(strings.Split(cfg.Addrs, ","), defaultCfg)
}
