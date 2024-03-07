package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Astemirdum/library-service/gateway/internal/model"

	"github.com/IBM/sarama"
)

type Enqueuer interface {
	Enqueue(topic string, v any) error
	EnqueueV2(ctx context.Context, fn func(ctx context.Context, userName string, stars int) (int, error), req model.RatingMsg)
}

func NewEnqueuer(producer sarama.SyncProducer) Enqueuer {
	return &enqueuerImpl{
		producer: producer,
	}
}

type enqueuerImpl struct {
	producer sarama.SyncProducer
}

func (q *enqueuerImpl) Enqueue(topic string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(data)}
	if _, _, err = q.producer.SendMessage(msg); err != nil {
		return err
	}
	return nil
}

func (q *enqueuerImpl) EnqueueV2(_ context.Context, fn func(ctx context.Context, userName string, stars int) (int, error), req model.RatingMsg) {
	go func() {
		for i := 0; i < 10; i++ {
			code, err := fn(context.Background(), req.Name, req.Stars)
			fmt.Println("EnqueueV2", code, err)
			if err == nil {
				return
			}
			time.Sleep(time.Second * 10)
		}
	}()
}
