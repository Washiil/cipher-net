package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly"
)

type player struct {
	uuid   string
	twitch string
}

func main() {
	region := flag.String("region", "na", "Leaderboard to Scrape")
	outputFile := flag.String("output", "data.db", "Path to the output data file")
	speed := flag.Float64("speed", 5, "How many pages to scrape per minute")
	verbose := flag.Bool("verbose", false, "Enable verbose output for debugging")

	flag.Parse()

	fmt.Println("--- Neural Theft Scraper ---")
	fmt.Printf(" > Region: %s\n", *region)
	fmt.Printf(" > Output File: %s\n", *outputFile)
	fmt.Printf(" > %.2f pages per min \n", *speed)

	if *verbose {
		fmt.Println(" > Verbose mode: ON")
	} else {
		fmt.Println(" > Verbose mode: OFF")
	}

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	c.OnHTML("td a[href*='/valorant/profile']", func(e *colly.HTMLElement) {
		username := e.ChildText("span.v3-trnign .max-w-full.truncate")
		tag := e.ChildText("span.v3-trnign__discriminator")

		twitchURL := e.DOM.Parent().Parent().Find(`a[aria-label="Visit twitch profile"]`).AttrOr("href", "")
		twitchName := ""
		if twitchURL != "" {
			parts := strings.Split(twitchURL, "/")
			twitchName = parts[len(parts)-1]
		}

		fmt.Printf("Username: %s\n", username)
		fmt.Printf("Tag: %s\n", tag)
		fmt.Printf("Twitch: %s\n", twitchName)
	})

	err := c.Visit("https://tracker.gg/valorant/leaderboards/ranked/all/default?platform=pc&region=na&act=ac12e9b3-47e6-9599-8fa1-0bb473e5efc7&page=1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n--- Execution Complete ---")
}
