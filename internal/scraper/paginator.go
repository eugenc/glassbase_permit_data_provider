package scraper

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/echayko/glassbase_permit_data_provider/internal/fetcher"
	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
	"github.com/tidwall/gjson"
)

type Page struct {
	Body         string
	NetworkCalls []fetcher.NetworkCall
	PageNum      int
	FetchURL     string
}

// Paginator iterates pages according to ConnectorConfig.Pagination.
type Paginator struct {
	config      *generator.ConnectorConfig
	baseURL     string
	fetcher     fetcher.Fetcher
	currentPage int
	done        bool
	pendingURL  string
	lastCursor  string
}

func NewPaginator(config *generator.ConnectorConfig, baseURL string, f fetcher.Fetcher) *Paginator {
	return &Paginator{
		config:  config,
		baseURL: baseURL,
		fetcher: f,
	}
}

func maxPagesConfig(p *generator.ConnectorConfig) int {
	if p.Pagination.MaxPages > 0 {
		return p.Pagination.MaxPages
	}
	return 100
}

// Next fetches the next page. Returns ok=false when pagination is exhausted.
func (p *Paginator) Next(ctx context.Context) (*Page, bool, error) {
	if p.done {
		return nil, false, nil
	}

	maxP := maxPagesConfig(p.config)
	if p.currentPage >= maxP {
		p.done = true
		return nil, false, nil
	}

	if p.currentPage > 0 && p.config.RateLimit.DelayBetweenRequestsMs > 0 {
		time.Sleep(time.Duration(p.config.RateLimit.DelayBetweenRequestsMs) * time.Millisecond)
	}

	fetchURL := p.pendingURL
	if fetchURL == "" {
		fetchURL = p.buildURL()
	} else {
		p.pendingURL = ""
	}

	result, err := p.fetcher.Fetch(ctx, fetchURL)
	if err != nil {
		return nil, false, err
	}

	p.currentPage++
	pageNum := p.currentPage

	switch p.config.Pagination.Type {
	case "none":
		p.done = true

	case "cursor":
		path := p.config.Pagination.CursorJSONPath
		if path == "" {
			p.done = true
			break
		}
		nextCursor := gjson.Get(result.Body, path).String()
		if nextCursor == "" {
			p.done = true
		} else if nextCursor == p.lastCursor {
			p.done = true
		} else {
			p.lastCursor = nextCursor
		}

	case "next_button":
		p.resolveNextButtonURL(result.Body)
		if p.pendingURL == "" {
			p.done = true
		}

	default:
		// page_param, offset: engine stops on empty records; paginator only stops at max pages
	}

	return &Page{
		Body:         result.Body,
		NetworkCalls: result.NetworkCalls,
		PageNum:      pageNum,
		FetchURL:     fetchURL,
	}, true, nil
}

func (p *Paginator) buildURL() string {
	u, err := url.Parse(p.baseURL)
	if err != nil {
		return p.baseURL
	}
	q := u.Query()

	switch p.config.Pagination.Type {
	case "page_param":
		param := p.config.Pagination.PageParam
		if param == "" {
			param = "page"
		}
		q.Set(param, fmt.Sprintf("%d", p.currentPage+1))
		u.RawQuery = q.Encode()
		return u.String()

	case "offset":
		offParam := p.config.Pagination.OffsetParam
		if offParam == "" {
			offParam = "offset"
		}
		size := p.config.Pagination.PageSize
		if size <= 0 {
			size = 20
		}
		q.Set(offParam, fmt.Sprintf("%d", p.currentPage*size))
		u.RawQuery = q.Encode()
		return u.String()

	case "cursor":
		param := p.config.Pagination.CursorParam
		if param == "" {
			param = "cursor"
		}
		if p.lastCursor != "" {
			q.Set(param, p.lastCursor)
		}
		u.RawQuery = q.Encode()
		return u.String()

	default:
		return p.baseURL
	}
}

func (p *Paginator) resolveNextButtonURL(html string) {
	sel := p.config.Pagination.NextSelector
	if sel == "" {
		p.pendingURL = ""
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		p.pendingURL = ""
		return
	}
	n := doc.Find(sel).First()
	if n.Length() == 0 {
		p.pendingURL = ""
		return
	}
	href, ok := n.Attr("href")
	if !ok || href == "" || href == "#" {
		p.pendingURL = ""
		return
	}
	base, err := url.Parse(p.baseURL)
	if err != nil {
		p.pendingURL = ""
		return
	}
	ref, err := url.Parse(href)
	if err != nil {
		p.pendingURL = ""
		return
	}
	p.pendingURL = base.ResolveReference(ref).String()
}
