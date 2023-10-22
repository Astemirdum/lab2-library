package kafka

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type Config struct {
	Addrs string `yaml:"addrs" envconfig:"KAFKA_ADDRS"`
}

const (
	LibraryTopic         = "library"
	RatingTopic          = "rating"
	LibraryConsumerGroup = "library"
	RatingConsumerGroup  = "rating"
)

func NewProducer(cfg Config) (sarama.SyncProducer, error) {
	defaultCfg := sarama.NewConfig()

	defaultCfg.Producer.RequiredAcks = sarama.WaitForAll
	defaultCfg.Producer.Return.Successes = true

	return sarama.NewSyncProducer(strings.Split(cfg.Addrs, ","), defaultCfg)
}

func NewConsumer(cfg Config, consumerGroup string) (sarama.ConsumerGroup, error) {
	defaultCfg := sarama.NewConfig()

	defaultCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	// defaultCfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	return sarama.NewConsumerGroup(strings.Split(cfg.Addrs, ","), consumerGroup, defaultCfg)

}

func Consume(client sarama.ConsumerGroup, handler sarama.ConsumerGroupHandler, topic string) {
	keepRunning := true
	ctx, cancel := context.WithCancel(context.Background())
	//consumer := Consumer{
	//	availableCountHandler: availableCountHandler,
	//	ready:                 make(chan bool),
	//}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, []string{topic}, handler); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()

	log.Println("Sarama consumer up and running!...")

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	for keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			log.Println("terminating: via signal")
			keepRunning = false
		}
	}
	cancel()
	wg.Wait()
	if err := client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}
