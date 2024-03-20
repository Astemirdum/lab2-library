package kafka

import (
	"fmt"
	"strings"

	"github.com/IBM/sarama"
)

type Config struct {
	Addrs string `yaml:"addrs" envconfig:"KAFKA_BROKERS"`
}

const (
	LibraryTopic = "library"
	RatingTopic  = "rating"
	StatsTopic   = "stats"

	LibraryConsumerGroup = "library"
	RatingConsumerGroup  = "rating"
	StatsConsumerGroup   = "stats"
)

func CreateTopics(cfg Config) error {
	defaultCfg := sarama.NewConfig()
	//defaultCfg.Version = sarama.V1_0_0_0
	brokerAddrs := strings.Split(cfg.Addrs, ",")
	admin, err := sarama.NewClusterAdmin(brokerAddrs, defaultCfg)
	if err != nil {
		return fmt.Errorf("creating cluster admin: %w", err)
	}
	topics, _ := admin.ListTopics() //nolint:errcheck
	fmt.Println("list topics", topics)
	defer admin.Close()
	if err := admin.CreateTopic("test", &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false); err != nil {
		return fmt.Errorf("create topics %w", err)
	}

	return nil
}
