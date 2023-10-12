package kafka

import (
	"github.com/IBM/sarama"
)

type Config struct {
	Addrs []string
}

func NewProducer(cfg Config) (sarama.SyncProducer, error) {
	defaultCfg := sarama.NewConfig()

	defaultCfg.Producer.RequiredAcks = sarama.WaitForAll
	defaultCfg.Producer.Return.Successes = true

	return sarama.NewSyncProducer(cfg.Addrs, defaultCfg)
}
