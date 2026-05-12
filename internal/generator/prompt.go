package generator

import "fmt"

// BuildPrompt constructs the user message for Claude.
func BuildPrompt(url, sourceType, pageContent string, networkCalls []string) string {
	networkSection := ""
	if len(networkCalls) > 0 {
		networkSection = fmt.Sprintf(`
The page made the following API/XHR calls during render. If any of these return permit data,
prefer using the API approach over HTML parsing:

%s
`, joinLines(networkCalls))
	}

	return fmt.Sprintf(`You are a web scraping expert analyzing a county building permit webpage.

URL: %s
Source Type: %s
%s

PAGE CONTENT (truncated to 8000 chars):
%s

---

Your task: Analyze this page and return a JSON object that tells a Go scraper how to extract
building permit records from it. The JSON must exactly follow this schema:

{
  "source_type": "html" | "spa" | "api",
  "api": {
    "endpoint": "<full URL if using API, optional>",
    "method": "GET",
    "headers": {},
    "body": ""
  },
  "extraction": {
    "record_selector": "<CSS selector for the repeating permit row/card — HTML/SPA only>",
    "records_path": "<JSON path to array of permits — API only, e.g. $.results>",
    "fields": [
      {
        "name": "<snake_case field name>",
        "selector": "<CSS selector relative to record container — HTML/SPA>",
        "json_path": "<dot-notation path within each record — API>",
        "attr": "<HTML attribute if not innerText, e.g. href>",
        "type": "text" | "date" | "number" | "url"
      }
    ]
  },
  "pagination": {
    "type": "none" | "page_param" | "offset" | "cursor" | "next_button",
    "page_param": "<URL query param name for page number>",
    "offset_param": "<URL query param name for offset when type is offset>",
    "page_size": <integer>,
    "cursor_param": "<query param for cursor token when type is cursor>",
    "cursor_json_path": "<gjson path to next cursor in API response, e.g. nextCursor>",
    "next_selector": "<CSS selector for next button>",
    "max_pages": <integer, default 100>
  },
  "dedup": {
    "unique_field": "<the field name that uniquely identifies a permit>"
  },
  "rate_limit": {
    "delay_between_requests_ms": 1000,
    "max_concurrent": 1
  }
}

Rules:
1. Include ALL fields visible on the page — permit number, address, date, type, status, contractor, owner, value, description, sq footage, etc.
2. The unique_field MUST be a field that appears in every record (typically permit number or ID).
3. If pagination is not visible, set type to "none".
4. If the page uses an API (detected from network calls above), set source_type to "api" and use records_path + json_path instead of CSS selectors; fill "api.endpoint" with the list/search URL if known.
5. Return ONLY the JSON object. No explanation, no markdown, no code fences.`,
		url, sourceType, networkSection, truncate(pageContent, 8000))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n... [truncated]"
}

func joinLines(lines []string) string {
	result := ""
	for _, l := range lines {
		result += "- " + l + "\n"
	}
	return result
}
