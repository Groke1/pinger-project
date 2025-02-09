package service

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-ping/ping"
	"pinger/configs"
	"pinger/models"
	"sync"
	"time"
)

type PingService interface {
	StartPing(ctx context.Context) (<-chan models.Ping, <-chan error)
	Stop()
}

type pingService struct {
	containersTimeout *time.Ticker
	pingTimeout       *time.Ticker
	workers           int
	containersIP      []string
	mx                sync.RWMutex
	cl                *client.Client
}

func NewPingService(cfg *configs.PingerConfig) (PingService, error) {
	cl, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &pingService{
		containersTimeout: time.NewTicker(cfg.ContainersTimeout),
		pingTimeout:       time.NewTicker(cfg.PingTimeout),
		workers:           cfg.Workers,
		cl:                cl,
	}, nil
}

func (p *pingService) StartPing(ctx context.Context) (<-chan models.Ping, <-chan error) {
	err := p.updateContainersIP(ctx)
	pingCh := make(chan models.Ping)
	errCh := make(chan error, 1)
	wgErr := &sync.WaitGroup{}
	wgPing := &sync.WaitGroup{}

	if err == nil {
		p.waitContainerTimeout(ctx, errCh, wgErr)
		p.waitPingTimeout(ctx, pingCh, errCh, wgPing, wgErr)

		go func() {
			wgErr.Wait()
			close(errCh)
		}()

		go func() {
			wgPing.Wait()
			close(pingCh)
		}()

		return pingCh, errCh
	}

	errCh <- err
	close(errCh)
	close(pingCh)
	return pingCh, errCh
}

func (p *pingService) Stop() {
	p.pingTimeout.Stop()
	p.containersTimeout.Stop()
	p.cl.Close()
}

func (p *pingService) waitContainerTimeout(ctx context.Context, errCh chan<- error, wgErr *sync.WaitGroup) {
	wgErr.Add(1)
	go func() {
		defer wgErr.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.containersTimeout.C:
				err := p.updateContainersIP(ctx)
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
}

func (p *pingService) updateContainersIP(ctx context.Context) error {

	containers, err := p.cl.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return err
	}

	var newIP []string
	for _, cont := range containers {
		for _, netSettings := range cont.NetworkSettings.Networks {
			newIP = append(newIP, netSettings.IPAddress)
		}
	}
	p.mx.Lock()
	p.containersIP = newIP
	p.mx.Unlock()
	return nil
}

func (p *pingService) waitPingTimeout(ctx context.Context, pingCh chan<- models.Ping, errCh chan<- error, wgPing, wgErr *sync.WaitGroup) {
	wgErr.Add(1)
	wgPing.Add(1)
	go func() {
		defer wgPing.Done()
		defer wgErr.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.pingTimeout.C:
				p.mx.RLock()
				ipCh := make(chan string, len(p.containersIP))
				p.mx.RUnlock()
				p.addIPInChan(ctx, ipCh)
				p.sendPing(ctx, ipCh, pingCh, errCh, wgPing, wgErr)
			}
		}
	}()
}

func (p *pingService) addIPInChan(ctx context.Context, ipCh chan<- string) {
	go func() {
		defer close(ipCh)
		p.mx.RLock()
		defer p.mx.RUnlock()
		for _, ip := range p.containersIP {
			select {
			case <-ctx.Done():
				return
			case ipCh <- ip:
			}
		}

	}()
}

func (p *pingService) sendPing(ctx context.Context, ipCh <-chan string, pingCh chan<- models.Ping, errCh chan<- error, wgPing, wgErr *sync.WaitGroup) {
	for worker := 0; worker < p.workers; worker++ {
		wgErr.Add(1)
		wgPing.Add(1)
		go func() {
			defer wgPing.Done()
			defer wgErr.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case ip, ok := <-ipCh:
					if !ok {
						return
					}
					pinger, err := ping.NewPinger(ip)
					if err != nil {
						select {
						case <-ctx.Done():
							return
						case errCh <- err:
						}
						continue
					}
					pinger.Count = 1

					pingTime := time.Now()
					err = pinger.Run()

					if err != nil {
						select {
						case <-ctx.Done():
							return
						case errCh <- err:
						}
						continue
					}
					stats := pinger.Statistics()
					newPing := models.Ping{
						IP:          ip,
						Duration:    int(stats.AvgRtt.Microseconds()),
						TimeAttempt: pingTime,
					}
					select {
					case <-ctx.Done():
						return
					case pingCh <- newPing:
					}

				}
			}

		}()
	}
}
