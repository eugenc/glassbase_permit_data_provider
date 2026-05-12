package runner

import (
	"context"
	"log"

	"github.com/echayko/glassbase_permit_data_provider/config"
	"github.com/echayko/glassbase_permit_data_provider/internal/monitor"
	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron    *cron.Cron
	workers *WorkerPool
	pool    *pgxpool.Pool
	cfg     *config.Config
}

func NewScheduler(pool *pgxpool.Pool, cfg *config.Config) *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		workers: NewWorkerPool(pool, cfg.MaxConcurrent),
		pool:    pool,
		cfg:     cfg,
	}
}

// Start registers scheduled jobs and starts the scheduler.
func (s *Scheduler) Start() {
	if _, err := s.cron.AddFunc("0 0 6 * * MON", func() {
		log.Println("cron: weekly permit scrape triggered")
		ctx := context.Background()
		s.workers.RunAll(ctx)
	}); err != nil {
		log.Fatalf("cron: failed to register weekly job: %v", err)
	}

	if _, err := s.cron.AddFunc("0 0 8 * * *", func() {
		log.Println("monitor: daily health probe starting")
		ctx := context.Background()
		probe := monitor.NewHealthProbe(s.pool)
		results := probe.ProbeAll(ctx)

		for _, r := range results {
			if !r.Healthy {
				monitor.SendAlert("warning", r.CountyID, "URL probe failed", r.Reason)
			}
		}
	}); err != nil {
		log.Fatalf("cron: failed to register health probe: %v", err)
	}

	if _, err := s.cron.AddFunc("0 0 9 * * *", func() {
		log.Println("monitor: zero-record check starting")
		ctx := context.Background()

		zeros, err := monitor.FindZeroCounties(ctx, s.pool, 2)
		if err != nil {
			log.Printf("zerocheck: %v", err)
			return
		}

		store := registry.NewStore(s.pool)
		for _, z := range zeros {
			monitor.SendAlert("warning", z.CountyID, "zero records on last 2 runs", "")
			_ = store.SetStatus(ctx, z.CountyID, "broken")
		}

		repairer := monitor.NewRepairer(s.pool, s.cfg.AnthropicAPIKey, s.cfg.AnthropicModel)
		repairer.RepairBroken(ctx)
	}); err != nil {
		log.Fatalf("cron: failed to register zero-record job: %v", err)
	}

	s.cron.Start()
	log.Println("cron: scheduler started — weekly Mon 06:00 UTC; health 08:00 UTC; zero-check 09:00 UTC")

	if s.cfg.RunNowOnStartup {
		go s.RunNow()
	}
}

// RunNow triggers an immediate full batch scrape.
func (s *Scheduler) RunNow() {
	log.Println("runner: manual trigger")
	ctx := context.Background()
	s.workers.RunAll(ctx)
}

// RunCounty triggers an immediate scrape for a single county (async).
func (s *Scheduler) RunCounty(countyID string) error {
	return s.workers.RunCounty(countyID)
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}
