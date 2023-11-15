package handler

import (
	"context"

	"encoding/json"

	"github.com/Astemirdum/library-service/library/internal/model"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type availableCount func(ctx context.Context, libraryID, bookID int, isReturn bool) error

type Consumer struct {
	availableCountHandler availableCount
	log                   *zap.Logger
	ready                 chan bool
}

func NewConsumer(availableCount availableCount, log *zap.Logger) *Consumer {
	return &Consumer{
		availableCountHandler: availableCount,
		log:                   log.Named("consumer"),
		ready:                 make(chan bool),
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
			var req model.AvailableCountRequest
			if err := json.Unmarshal(message.Value, &req); err != nil {
				consumer.log.Error("", zap.Error(err))
				session.MarkMessage(message, "")
				continue
			}

			if err := consumer.availableCountHandler(context.Background(), req.LibraryID, req.BookID, req.IsReturn); err != nil {
				consumer.log.Error("consumer.availableCountHandler", zap.Error(err))
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
