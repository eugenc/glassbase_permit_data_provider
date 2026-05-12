package monitor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/echayko/glassbase_permit_data_provider/internal/registry"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthProbe struct {
	pool   *pgxpool.Pool
	client *http.Client
}

func NewHealthProbe(pool *pgxpool.Pool) *HealthProbe {
	return &HealthProbe{
		pool:   pool,
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

type ProbeResult struct {
	CountyID   string
	URL        string
	StatusCode int
	Healthy    bool
	Reason     string
}

func (h *HealthProbe) ProbeAll(ctx context.Context) []ProbeResult {
	store := registry.NewStore(h.pool)
	counties, err := store.GetAll(ctx)
	if err != nil {
		log.Printf("monitor: failed to load counties: %v", err)
		return nil
	}

	var results []ProbeResult
	for _, county := range counties {
		result := h.probeOne(ctx, &county)
		results = append(results, result)
		if !result.Healthy {
			log.Printf("monitor: [%s] UNHEALTHY — %s", county.CountyID, result.Reason)
		}
		if !result.Healthy && strings.Contains(strings.ToLower(result.Reason), "login") {
			_ = store.SetStatus(ctx, county.CountyID, "paused")
		}
	}

	return results
}

func (h *HealthProbe) probeOne(ctx context.Context, county *registry.CountyConnector) ProbeResult {
	result := ProbeResult{
		CountyID: county.CountyID,
		URL:      county.URL,
		Healthy:  true,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, county.URL, nil)
	if err != nil {
		result.Healthy = false
		result.Reason = fmt.Sprintf("bad URL: %v", err)
		return result
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := h.client.Do(req)
	if err != nil {
		result.Healthy = false
		result.Reason = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	if resp.StatusCode >= 400 {
		result.Healthy = false
		result.Reason = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result
	}

	finalURL := resp.Request.URL.String()
	if finalURL != county.URL {
		low := strings.ToLower(finalURL)
		if strings.Contains(low, "login") || strings.Contains(low, "signin") ||
			strings.Contains(low, "auth") || strings.Contains(low, "sso") {
			result.Healthy = false
			result.Reason = fmt.Sprintf("redirected to login: %s", finalURL)
			return result
		}
	}

	return result
}
