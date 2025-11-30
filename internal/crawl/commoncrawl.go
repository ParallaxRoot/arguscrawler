package commoncrawl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CommonCrawlSource struct {
	client *http.Client
}

func New() *CommonCrawlSource {
	return &CommonCrawlSource{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

type ccCollection struct {
	ID     string `json:"id"`
	CDXAPI string `json:"cdx-api"`
}

// Função principal
func (s *CommonCrawlSource) Fetch(domain string) ([]string, error) {
	ctx := context.Background()

	req, _ := http.NewRequestWithContext(ctx, "GET",
		"https://index.commoncrawl.org/collinfo.json", nil)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching collections: %w", err)
	}
	defer resp.Body.Close()

	var collections []ccCollection
	if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
		return nil, fmt.Errorf("decode collections: %w", err)
	}

	seen := map[string]struct{}{}

	if len(collections) > 3 {
		collections = collections[:3]
	}

	for _, col := range collections {

		indexURL := fmt.Sprintf(
			"%s?url=*.%s&output=json&matchType=domain",
			col.CDXAPI,
			url.QueryEscape(domain),
		)

		req2, _ := http.NewRequestWithContext(ctx, "GET", indexURL, nil)
		resp2, err := s.client.Do(req2)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(resp2.Body)
		for scanner.Scan() {
			line := scanner.Text()

			var record map[string]interface{}
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				continue
			}

			rawURL, _ := record["url"].(string)
			host := extractHost(rawURL)

			if host == "" {
				continue
			}

			if host == domain || strings.HasSuffix(host, "."+domain) {
				seen[host] = struct{}{}
			}
		}

		resp2.Body.Close()
	}

	out := make([]string, 0, len(seen))
	for h := range seen {
		out = append(out, h)
	}

	return out, nil
}

func extractHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}
