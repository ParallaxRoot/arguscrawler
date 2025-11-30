package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ParallaxRoot/arguscrawler/internal/config"
	"github.com/ParallaxRoot/arguscrawler/internal/logger"
	"github.com/ParallaxRoot/arguscrawler/internal/passive"
)

func main() {
	printBanner()

	domain := flag.String("d", "", "Domain to crawl (e.g. example.com)")
	flag.Parse()

	if *domain == "" {
		fmt.Println("[!] Provide -d <domain>")
		os.Exit(1)
	}

	log := logger.New()

	cfg := config.Config{
		Domain: *domain,
	}

	source := passive.NewCommonCrawlSource(log)

	results, err := source.Enum(context.Background(), cfg.Domain)
	if err != nil {
		log.Errorf("CommonCrawl error: %v", err)
		os.Exit(1)
	}

	log.Infof("Found %d subdomains:", len(results))
	for _, s := range results {
		fmt.Println(s)
	}
}

func printBanner() {
	fmt.Println(`
   ╔══════════════════════════════════════════════╗
   ║              ArgusCrawler v0.1              ║
   ║                CommonCrawl Mode             ║
   ╚══════════════════════════════════════════════╝`)
}
