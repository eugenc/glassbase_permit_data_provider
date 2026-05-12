package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type SPAFetcher struct{}

func (f *SPAFetcher) Fetch(ctx context.Context, url string) (*FetchResult, error) {
	l := launcher.New().Headless(true).Leakless(false)
	if bin := os.Getenv("CHROME_BIN"); bin != "" {
		l = l.Bin(bin)
	}
	u, err := l.Launch()
	if err != nil {
		return nil, err
	}

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage()

	navCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	page = page.Context(navCtx)

	router := page.HijackRequests()
	var (
		networkCalls []NetworkCall
		mu           sync.Mutex
	)
	router.MustAdd("*", func(h *rod.Hijack) {
		h.MustLoadResponse()
		mu.Lock()
		defer mu.Unlock()
		reqURL := ""
		if h.Request != nil && h.Request.URL() != nil {
			reqURL = h.Request.URL().String()
		}
		method := ""
		if h.Request != nil {
			method = h.Request.Method()
		}
		body := ""
		if h.Request != nil {
			body = h.Request.Body()
		}
		resp := ""
		if h.Response != nil {
			resp = truncateRunes(h.Response.Body(), 50000)
		}
		// Capture XHR/fetch style traffic (skip main document duplicates minimally)
		networkCalls = append(networkCalls, NetworkCall{
			URL:      reqURL,
			Method:   method,
			Body:     body,
			Response: resp,
		})
	})
	go router.Run()
	defer func() { _ = router.Stop() }()

	if err := page.Navigate(url); err != nil {
		return nil, err
	}
	page.MustWaitLoad()

	select {
	case <-time.After(2 * time.Second):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	html, err := page.HTML()
	if err != nil {
		return nil, err
	}

	if len(trimSpace(html)) == 0 {
		return nil, fmt.Errorf("spa fetch %s: empty body", url)
	}

	return &FetchResult{
		Body:         html,
		StatusCode:   http.StatusOK,
		SourceType:   "spa",
		NetworkCalls: networkCalls,
	}, nil
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\n' || s[0] == '\t' || s[0] == '\r') {
		s = s[1:]
	}
	return s
}

func truncateRunes(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}
