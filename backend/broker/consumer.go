package broker

import (
	"backend/configs"
	"backend/models"
	"backend/service"
	"bytes"
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
	"sync"
)

type Consumer interface {
	StartGetPings(ctx context.Context) <-chan error
	HandleEvents(ctx context.Context)
	Close() error
}

type consumer struct {
	serv service.Service
	cons *kafka.Consumer
	cfg  *configs.ConsumerConfig
	log  *logrus.Logger
}

func NewConsumer(cfg *configs.ConsumerConfig, serv service.Service, log *logrus.Logger) (Consumer, error) {
	cons, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Host + ":" + cfg.Port,
		"group.id":          cfg.GroupID,
		"auto.offset.reset": cfg.AutoOffsetReset,
	})
	if err != nil {
		return nil, err
	}
	return &consumer{serv: serv, cons: cons, cfg: cfg, log: log}, nil
}

func (c *consumer) Close() error {
	return c.cons.Close()
}

func (c *consumer) HandleEvents(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-c.cons.Events():
				if !ok {
					return
				}
				switch e := event.(type) {
				case kafka.Error:
					c.log.Error(e)
				}
			}
		}
	}()
}

func (c *consumer) StartGetPings(ctx context.Context) <-chan error {
	errorCh := make(chan error, 1)
	err := c.createTopic(ctx)
	if err != nil {
		errorCh <- err
		close(errorCh)
		return errorCh
	}

	err = c.cons.Subscribe(c.cfg.PingTopic, nil)
	if err != nil {
		errorCh <- err
		close(errorCh)
		return errorCh
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := c.cons.ReadMessage(-1)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case errorCh <- err:
					continue
				}
			}
			var ping models.Ping
			if err := json.NewDecoder(bytes.NewReader(msg.Value)).Decode(&ping); err != nil {
				select {
				case <-ctx.Done():
					return
				case errorCh <- err:
					continue
				}
			}

			c.log.Infof("get ping: %s", ping)

			err = c.serv.AddPing(ctx, ping)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case errorCh <- err:
				}
			}

		}
	}()

	go func() {
		wg.Wait()
		close(errorCh)
	}()

	return errorCh

}

func (c *consumer) createTopic(ctx context.Context) error {
	adminClient, err := kafka.NewAdminClientFromConsumer(c.cons)
	if err != nil {
		return err
	}
	defer adminClient.Close()

	listTopics := []kafka.TopicSpecification{
		{
			Topic:             c.cfg.PingTopic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}
	_, err = adminClient.CreateTopics(ctx, listTopics)
	if err != nil {
		return err
	}

	return nil
}
