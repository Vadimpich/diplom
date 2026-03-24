package decision

import (
	"context"
	"log"
	"time"
)

type Relay struct {
	repo         Repository
	service      *Service
	pollInterval time.Duration
	logger       *log.Logger
}

func NewRelay(repo Repository, service *Service, pollInterval time.Duration) *Relay {
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	return &Relay{
		repo:         repo,
		service:      service,
		pollInterval: pollInterval,
		logger:       log.Default(),
	}
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for {
		if err := r.runOnce(ctx); err != nil && ctx.Err() == nil {
			r.logger.Printf("decision relay error=%v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *Relay) runOnce(ctx context.Context) error {
	if r.repo == nil || r.service == nil {
		return nil
	}
	snapshots, err := r.repo.ListPendingSnapshots(ctx)
	if err != nil {
		return err
	}
	for _, snapshot := range snapshots {
		if err := r.service.DeliverPending(ctx, snapshot); err != nil {
			return err
		}
	}
	return nil
}
