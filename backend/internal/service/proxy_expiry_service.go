package service

import (
	"context"
	"log"
	"sync"
	"time"
)

// ProxyExpiryService periodically scans for expired proxies and re-routes
// bound accounts to fallback or direct connections.
type ProxyExpiryService struct {
	proxyRepo ProxyRepository
	interval  time.Duration
	stopCh    chan struct{}
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

func NewProxyExpiryService(proxyRepo ProxyRepository, interval time.Duration) *ProxyExpiryService {
	return &ProxyExpiryService{proxyRepo: proxyRepo, interval: interval, stopCh: make(chan struct{})}
}

func (svc *ProxyExpiryService) Start() {
	if svc == nil || svc.proxyRepo == nil || svc.interval <= 0 {
		return
	}
	svc.wg.Add(1)
	go func() {
		defer svc.wg.Done()
		tick := time.NewTicker(svc.interval)
		defer tick.Stop()
		svc.sweep()
		for {
			select {
			case <-tick.C:
				svc.sweep()
			case <-svc.stopCh:
				return
			}
		}
	}()
}

func (svc *ProxyExpiryService) Stop() {
	if svc == nil {
		return
	}
	svc.stopOnce.Do(func() { close(svc.stopCh) })
	svc.wg.Wait()
}

func (svc *ProxyExpiryService) sweep() {
	deadline, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	affected, sweepErr := svc.proxyRepo.SweepExpiredProxies(deadline, time.Now())
	if sweepErr != nil {
		log.Printf("[ProxyExpiry] failed to sweep expired proxies: %v", sweepErr)
		return
	}
	if affected > 0 {
		log.Printf("[ProxyExpiry] migrated %d accounts away from expired proxies", affected)
	}
}
