package handler

import (
	"encoding/json"

	"github.com/Astemirdum/library-service/backend/pkg/kafka"
	"github.com/IBM/sarama"
)

type statsLog struct {
	producer sarama.AsyncProducer
	topic    string
}

type StatsLog interface {
	Log(sl kafka.EventStats) error
}

func NewStatsLog(producer sarama.AsyncProducer, topic string) *statsLog {
	return &statsLog{
		producer: producer,
		topic:    topic,
	}
}

func (l *statsLog) Log(sl kafka.EventStats) error {
	if l == nil {
		return nil
	}
	data, err := json.Marshal(sl)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{Topic: l.topic, Value: sarama.StringEncoder(data)}
	l.producer.Input() <- msg
	return nil
}
