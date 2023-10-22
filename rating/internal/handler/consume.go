package handler

import (
	"context"
	"encoding/json"

	"github.com/Astemirdum/library-service/rating/internal/model"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type rating func(ctx context.Context, name string, stars int) error

type Consumer struct {
	ratingHandler rating
	log           *zap.Logger
	ready         chan bool
}

func newConsumer(rating rating, log *zap.Logger) *Consumer {
	return &Consumer{
		ratingHandler: rating,
		log:           log.Named("consumer"),
		ready:         make(chan bool),
	}
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
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
			var req model.RatingMsg
			if err := json.Unmarshal(message.Value, &req); err != nil {
				consumer.log.Error("", zap.Error(err))
				session.MarkMessage(message, "")
				continue
			}

			if err := consumer.ratingHandler(context.Background(), req.Name, req.Stars); err != nil {
				consumer.log.Error("consumer.ratingHandler", zap.Error(err))
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
