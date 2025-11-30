package crawler

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

func NewCommonCrawl(log *logger.Logger) *CommonCrawl {
	return &CommonCrawl{
		log: log,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *CommonCrawl) Name() string {
	return "commoncrawl"
}

type ccCollection struct {
	ID     string `json:"id"`
	CDXAPI string `json:"cdx-api"`
}

// ---------------------------------------------------------------------------
//                              Core Crawler
// ---------------------------------------------------------------------------

func (c *CommonCrawl) Crawl(ctx context.Context, domain string) ([]string, error) {
	c.log.Infof("[+] Running crawler: %s", c.Name())

	// 1) Buscar lista de coleções completas 
	cols, err := c.fetchCollections(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})

	// 2) Consultar cada coleção em paralelo
	for _, col := range cols {
		c.queryCollection(ctx, col, domain, seen)
	}

	// 3) Converter para slice final
	out := make([]string, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}

	c.log.Infof("[commoncrawl] total collected: %d", len(out))
	return out, nil
}

// ---------------------------------------------------------------------------
//                          Fetch Collections
// ---------------------------------------------------------------------------

func (c *CommonCrawl) fetchCollections(ctx context.Context) ([]ccCollection, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET",
		"https://index.commoncrawl.org/collinfo.json", nil)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch collections: %w", err)
	}
	defer resp.Body.Close()

	var cols []ccCollection
	if err := json.NewDecoder(resp.Body).Decode(&cols); err != nil {
		return nil, fmt.Errorf("decode collections: %w", err)
	}

	// NÃO VAMOS LIMITAR — usamos todas as coleções
	c.log.Infof("[CommonCrawl] total collections: %d", len(cols))

	return cols, nil
}

// ---------------------------------------------------------------------------
//                             Query One Collection
// ---------------------------------------------------------------------------

func (c *CommonCrawl) queryCollection(
	ctx context.Context,
	col ccCollection,
	domain string,
	seen map[string]struct{},
) {

	indexURL := fmt.Sprintf(
		"%s?url=*.%s&matchType=domain&output=json",
		col.CDXAPI,
		url.QueryEscape(domain),
	)

	c.log.Infof("Querying CC index: %s", indexURL)

	req, _ := http.NewRequestWithContext(ctx, "GET", indexURL, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Errorf("[!] error %s: %v", col.ID, err)
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.Contains(line, domain) {
			continue
		}

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
}

// ---------------------------------------------------------------------------
//                           Extract Host
// ---------------------------------------------------------------------------

func extractHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}
