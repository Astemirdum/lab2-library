package handler

import (
	"encoding/json"

	"github.com/IBM/sarama"
)

type Enqueuer interface {
	Enqueue(topic string, v any) error
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
	if q.producer == nil {
		return nil
	}
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
