package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
	"pinger/configs"
	"pinger/models"
	"sync"
)

type Producer interface {
	SendPings(ctx context.Context, pingsCh <-chan models.Ping) <-chan error
	StartEventListener()
	Close()
}

type producer struct {
	prod *kafka.Producer
	cfg  *configs.ProducerConfig
	log  *logrus.Logger
}

func NewProducer(cfg *configs.ProducerConfig, log *logrus.Logger) (Producer, error) {
	prod, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
	})
	if err != nil {
		return nil, err
	}
	return &producer{
		prod: prod,
		cfg:  cfg,
		log:  log,
	}, nil
}

func (p *producer) StartEventListener() {
	go func() {
		for event := range p.prod.Events() {
			switch ev := event.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					p.log.Errorf("message not delivered: %s", ev.TopicPartition)
				}
			}
		}
	}()
}

func (p *producer) SendPings(ctx context.Context, pingCh <-chan models.Ping) <-chan error {
	errCh := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		for ping := range pingCh {
			err := p.sendPing(ping)
			if err != nil {
				select {
				case <-ctx.Done():
				case errCh <- err:
				}
			}

		}

	}()
	go func() {
		wg.Wait()
		close(errCh)
	}()
	return errCh
}

func (p *producer) sendPing(ping models.Ping) error {
	p.log.Infof("sending ping %v", ping)
	data, err := json.Marshal(ping)
	if err != nil {
		return err
	}

	p.prod.ProduceChannel() <- &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.cfg.PingTopic, Partition: kafka.PartitionAny},
		Value:          data,
	}

	return nil
}

func (p *producer) Close() {
	p.prod.Close()
}
