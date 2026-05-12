package runner

import (
	"context"
	"log"
	"sync"

	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/echayko/glassbase_permit_data_provider/internal/scraper"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultMaxConcurrent = 5

type WorkerPool struct {
	pool   *pgxpool.Pool
	engine *scraper.Engine
	runlog *RunLog
	sem    chan struct{}
}

func NewWorkerPool(pool *pgxpool.Pool, maxConcurrent int) *WorkerPool {
	if maxConcurrent <= 0 {
		maxConcurrent = defaultMaxConcurrent
	}
	return &WorkerPool{
		pool:   pool,
		engine: scraper.NewEngine(pool),
		runlog: NewRunLog(pool),
		sem:    make(chan struct{}, maxConcurrent),
	}
}

// RunAll scrapes all active counties and blocks until all are done.
func (w *WorkerPool) RunAll(ctx context.Context) {
	store := registry.NewStore(w.pool)
	counties, err := store.GetByStatus(ctx, "active")
	if err != nil {
		log.Printf("runner: failed to load active counties: %v", err)
		return
	}

	log.Printf("runner: starting batch — %d counties to scrape", len(counties))

	var wg sync.WaitGroup

	for _, county := range counties {
		county := county

		wg.Add(1)
		w.sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-w.sem }()

			w.scrapeOne(ctx, &county)
		}()
	}

	wg.Wait()
	log.Printf("runner: batch complete")
}

func (w *WorkerPool) scrapeOne(ctx context.Context, county *registry.CountyConnector) {
	runID, err := w.runlog.Start(ctx, county.CountyID)
	if err != nil {
		log.Printf("[%s] failed to start run log: %v", county.CountyID, err)
		return
	}

	log.Printf("[%s] scraping...", county.CountyID)

	result, err := w.engine.ScrapeCounty(ctx, county)

	if err != nil {
		log.Printf("[%s] FAILED: %v", county.CountyID, err)
		_ = w.runlog.Finish(ctx, runID, "failed", 0, 0, err.Error())

		store := registry.NewStore(w.pool)
		_ = store.SetStatus(ctx, county.CountyID, "broken")
		return
	}

	log.Printf("[%s] OK: %d found, %d inserted in %s",
		county.CountyID, result.RecordsFound, result.RecordsInserted, result.Duration)

	_ = w.runlog.Finish(ctx, runID, "success",
		result.RecordsFound, result.RecordsInserted, "")
}
