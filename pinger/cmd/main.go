package main

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"os/signal"
	"pinger/broker"
	"pinger/configs"
	"pinger/service"
	"syscall"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := logrus.New()
	cfg, err := configs.LoadConfig("configs/config.yaml")

	if err != nil {
		log.Error(err)
		return
	}

	serv, err := service.NewPingService(&cfg.PingerConfig)
	if err != nil {
		log.Error(err)
		return
	}
	defer serv.Stop()

	pingCh, errCh := serv.StartPing(ctx)
	logErrors(log, errCh)

	prod, err := broker.NewProducer(&cfg.ProducerConfig, log)
	if err != nil {
		log.Error(err)
		return
	}
	defer prod.Close()

	prod.StartEventListener()

	prodErrCh := prod.SendPings(ctx, pingCh)
	logErrors(log, prodErrCh)

	<-ctx.Done()
	log.Info("server shutdown")

}

func logErrors(log *logrus.Logger, errCh <-chan error) {
	go func() {
		for err := range errCh {
			log.Error(err)
		}
	}()
}
