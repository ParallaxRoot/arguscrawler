package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ParallaxRoot/arguscrawler/internal/crawl"
)

func main() {
	domain := flag.String("d", "", "Domain to crawl (example.com)")
	flag.Parse()

	if *domain == "" {
		fmt.Println("Use: arguscrawler -d example.com")
		os.Exit(1)
	}

	fmt.Println("ArgusCrawler — CommonCrawl extractor")
	fmt.Println("Domain:", *domain)

	cc := commoncrawl.New()

	subs, err := cc.Fetch(*domain)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Printf("\nFound %d results:\n", len(subs))
	for _, s := range subs {
		fmt.Println(" -", s)
	}
}
