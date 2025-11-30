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

	"github.com/ParallaxRoot/arguscrawler/internal/logger"
)

type CommonCrawl struct {
	log    *logger.Logger
	client *http.Client
}

func New(log *logger.Logger) *CommonCrawl {
	return &CommonCrawl{
		log: log,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type ccCollection struct {
	ID     string `json:"id"`
	CDXAPI string `json:"cdx-api"`
}

func (c *CommonCrawl) FetchCollections(ctx context.Context) ([]ccCollection, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET",
		"https://index.commoncrawl.org/collinfo.json", nil)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request collinfo: %w", err)
	}
	defer resp.Body.Close()

	var cols []ccCollection
	if err := json.NewDecoder(resp.Body).Decode(&cols); err != nil {
		return nil, fmt.Errorf("decode collinfo: %w", err)
	}

	return cols, nil
}

func (c *CommonCrawl) Enum(ctx context.Context, domain string) ([]string, error) {
	c.log.Info("[CommonCrawl] Fetching collections...")

	cols, err := c.FetchCollections(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})

	for i, col := range cols {
		if i >= 5 { // limita para não matar sua internet
			break
		}

		query := fmt.Sprintf("%s?url=*.%s&matchType=domain&output=json",
			col.CDXAPI,
			url.QueryEscape(domain),
		)

		c.log.Infof("Querying index: %s", col.ID)

		req, _ := http.NewRequestWithContext(ctx, "GET", query, nil)
		resp, err := c.client.Do(req)
		if err != nil {
			c.log.Errorf("[!] error: %v", err)
			continue
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			var data map[string]interface{}
			if json.Unmarshal([]byte(line), &data) != nil {
				continue
			}

			rawURL, _ := data["url"].(string)
			host := extractHost(rawURL)
			if host == "" {
				continue
			}

			if strings.HasSuffix(host, "."+domain) {
				seen[host] = struct{}{}
			}
		}

		resp.Body.Close()
	}

	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}

	c.log.Infof("[CommonCrawl] found %d hosts", len(out))
	return out, nil
}

func extractHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}
