package main

import (
	"backend/broker"
	"backend/configs"
	"backend/handler"
	"backend/repository"
	"backend/repository/db/postgres"
	"backend/server"
	"backend/service"
	"context"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
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

	db, err := postgres.NewPostgres(&postgres.Config{
		Pass:    os.Getenv("DB_PASS"),
		User:    os.Getenv("DB_USER"),
		DBName:  os.Getenv("DB_NAME"),
		Host:    cfg.DBConfig.Host,
		Port:    cfg.DBConfig.Port,
		SSLMode: cfg.DBConfig.SSLMode,
	})

	if err != nil {
		log.Error(err)
		return
	}

	defer db.Close()

	repo := repository.NewRepository(db)

	serv := service.NewService(repo)
	defer serv.Close()

	hand := handler.NewHandler(serv, log)

	router := mux.NewRouter()
	hand.InitRoutes(router)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	})

	consumer, err := broker.NewConsumer(&cfg.ConsumerConfig, serv, log)
	if err != nil {
		log.Error(err)
		return
	}
	defer consumer.Close()

	servErrCh := serv.WaitBatchTimeout(ctx)

	logErrors(log, servErrCh)

	consumer.HandleEvents(ctx)
	errorCh := consumer.StartGetPings(ctx)

	logErrors(log, errorCh)

	httpHandler := c.Handler(router)
	httpServer := server.NewServer(&cfg.ServerConfig, httpHandler)
	go func() {
		defer cancel()
		if err := httpServer.Run(); err != nil {
			log.Error(err)
		}
	}()
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
