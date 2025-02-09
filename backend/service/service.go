package service

import (
	"backend/models"
	"backend/repository"
	"context"
	"sync"
	"time"
)

type Service interface {
	GetPings(ctx context.Context) ([]models.Ping, error)
	AddPing(ctx context.Context, ping models.Ping) error
	WaitBatchTimeout(ctx context.Context) <-chan error
	Close()
}

type batchPings struct {
	bufferPings []models.Ping
	maxSize     int
	timeout     *time.Ticker
	mx          sync.Mutex
}

func newBatchPings(maxSize int, timeout *time.Ticker) *batchPings {
	return &batchPings{
		maxSize: maxSize,
		timeout: timeout,
	}
}

func (b *batchPings) clear() {
	b.bufferPings = b.bufferPings[:0]
}

type service struct {
	repo  repository.Repository
	batch *batchPings
}

func NewService(repo repository.Repository) Service {
	return &service{
		repo:  repo,
		batch: newBatchPings(50, time.NewTicker(10*time.Second)),
	}
}

func (s *service) GetPings(ctx context.Context) ([]models.Ping, error) {
	return s.repo.GetPings(ctx)
}

func (s *service) AddPing(ctx context.Context, ping models.Ping) error {
	s.batch.mx.Lock()
	defer s.batch.mx.Unlock()
	if len(s.batch.bufferPings) == s.batch.maxSize {
		err := s.repo.AddPings(ctx, s.batch.bufferPings)
		s.batch.clear()
		return err
	}
	s.batch.bufferPings = append(s.batch.bufferPings, ping)
	return nil
}

func (s *service) WaitBatchTimeout(ctx context.Context) <-chan error {
	errCh := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.batch.timeout.C:
				s.batch.mx.Lock()
				err := s.repo.AddPings(ctx, s.batch.bufferPings)
				s.batch.clear()
				s.batch.mx.Unlock()
				if err != nil {
					select {
					case <-ctx.Done():
						return
					case errCh <- err:
					}
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

func (s *service) Close() {
	s.batch.timeout.Stop()
}
