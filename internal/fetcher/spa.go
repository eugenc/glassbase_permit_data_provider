package fetcher

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type SPAFetcher struct{}

func (f *SPAFetcher) Fetch(ctx context.Context, url string, _ *FetchContext) (*FetchResult, error) {
	log.Printf("fetcher/spa: launching browser headless=true")
	l := launcher.New().Headless(true).Leakless(false)
	if bin := os.Getenv("CHROME_BIN"); bin != "" {
		log.Printf("fetcher/spa: CHROME_BIN=%s", bin)
		l = l.Bin(bin)
	}
	u, err := l.Launch()
	if err != nil {
		return nil, err
	}
	log.Printf("fetcher/spa: browser websocket url obtained")

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()
	log.Printf("fetcher/spa: connected, new page")

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
		if err := h.LoadResponse(http.DefaultClient, true); err != nil {
			// Stop(), page timeout, or parent cancel while XHR is in flight — MustLoadResponse would panic.
			switch {
			case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
				h.Response.Fail(proto.NetworkErrorReasonAborted)
			default:
				h.Response.Fail(proto.NetworkErrorReasonFailed)
			}
			return
		}
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
	log.Printf("fetcher/spa: hijack router running")
	defer func() { _ = router.Stop() }()

	log.Printf("fetcher/spa: navigating url=%s", url)
	if err := page.Navigate(url); err != nil {
		return nil, err
	}
	log.Printf("fetcher/spa: waiting for load event")
	page.MustWaitLoad()

	log.Printf("fetcher/spa: settling 2s for XHR")
	select {
	case <-time.After(2 * time.Second):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	html, err := page.HTML()
	if err != nil {
		return nil, err
	}
	log.Printf("fetcher/spa: html read bytes=%d captured_network_calls=%d", len(html), len(networkCalls))

	if len(trimSpace(html)) == 0 {
		return nil, fmt.Errorf("spa fetch %s: empty body", url)
	}

	log.Printf("fetcher/spa: success")
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
