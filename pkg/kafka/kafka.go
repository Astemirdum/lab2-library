package kafka

type Config struct {
	Addrs string `yaml:"addrs" envconfig:"KAFKA_ADDRS"`
}

const (
	LibraryTopic = "library"
	RatingTopic  = "rating"
	StatsTopic   = "stats"

	LibraryConsumerGroup = "library"
	RatingConsumerGroup  = "rating"
	StatsConsumerGroup   = "stats"
)
