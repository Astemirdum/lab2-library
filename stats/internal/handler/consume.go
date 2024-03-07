package handler

import (
	"context"
	"encoding/json"

	"github.com/Astemirdum/library-service/pkg/kafka"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type stats func(ctx context.Context, event kafka.EventStats) error

type Consumer struct {
	statsHandler stats
	log          *zap.Logger
	ready        chan bool
}

func NewConsumer(stats stats, log *zap.Logger) *Consumer {
	return &Consumer{
		statsHandler: stats,
		log:          log.Named("consumer"),
		ready:        make(chan bool),
	}
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				consumer.log.Warn("message channel was closed")
				return nil
			}
			var event kafka.EventStats
			if err := json.Unmarshal(message.Value, &event); err != nil {
				consumer.log.Error("", zap.Error(err))
				session.MarkMessage(message, "")
				continue
			}

			if err := consumer.statsHandler(context.Background(), event); err != nil {
				consumer.log.Error("consumer.statsHandler", zap.Error(err))
				//session.MarkMessage(message, "")
				continue
			}

			consumer.log.Debug("Message claimed:", zap.String("value", string(message.Value)), zap.Time("timestamp", message.Timestamp), zap.String("topic", message.Topic))
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
