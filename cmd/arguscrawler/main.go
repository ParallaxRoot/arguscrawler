package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/ParallaxRoot/arguscrawler/internal/crawl"
	"github.com/ParallaxRoot/arguscrawler/internal/logger"
)

func main() {

	domain := flag.String("d", "", "Domain to crawl from CommonCrawl")
	flag.Parse()

	if *domain == "" {
		fmt.Println("You must pass -d <domain>")
		return
	}

	fmt.Println("ArgusCrawler — CommonCrawl extractor")
	fmt.Println("Domain:", *domain)
	fmt.Println()

	log := logger.New()
	cc := commoncrawl.New(log)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := cc.Enum(ctx, *domain)
	if err != nil {
		log.Errorf("Error: %v", err)
		return
	}

	fmt.Printf("Found %d results:\n", len(result))
	for _, r := range result {
		fmt.Println(r)
	}
}
