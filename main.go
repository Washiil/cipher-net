package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

// Config holds all the configuration for the application.
type Config struct {
	Region     string
	OutputFile string
	Speed      float64 // This is now requests per minute for the translator
	Token      string
	Verbose    bool
}

// scrapedPlayer is the raw data structure produced by the scraper.
type scrapedPlayer struct {
	Name   string
	Tag    string
	Twitch string
}

// TranslatedItem is the data structure after being processed by the translator.
type TranslatedItem struct {
	UUID   string
	Twitch string
	Name   string
	Tag    string
}

func loadConfig() Config {
	region := flag.String("region", "na", "Leaderboard to scrape")
	outputFile := flag.String("output", "data.db", "Path to the output data file")
	speed := flag.Float64("speed", 75, "Users to scrape per minute")
	token := flag.String("token", "", "API Token for UUID parsing")
	verbose := flag.Bool("verbose", false, "Enable verbose output for debugging")
	flag.Parse()

	// Load token from .env if not set
	if *token == "" {
		log.Println("No token specified, attempting to read environmental variables...")
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}
		envToken := os.Getenv("API_KEY")
		if envToken == "" {
			log.Fatal("API_KEY not set in .env")
		}
		*token = envToken
	}

	return Config{
		Region:     *region,
		OutputFile: *outputFile,
		Speed:      *speed,
		Token:      *token,
		Verbose:    *verbose,
	}
}

func logVerbose(enabled bool, format string, args ...interface{}) {
	if enabled {
		log.Printf(format, args...)
	}
}

func scrapeLeaderboard(cfg Config, scrapedPlayerChan chan<- scrapedPlayer) {
	defer close(scrapedPlayerChan)

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	c.OnHTML("td a[href*='/valorant/profile']", func(e *colly.HTMLElement) {
		raw_name := e.ChildText("span.v3-trnign .max-w-full.truncate")
		username := strings.TrimSuffix(strings.TrimSpace(raw_name), "#")

		rawTag := e.ChildText("span.v3-trnign__discriminator")
		tag := strings.TrimPrefix(strings.TrimSpace(rawTag), "#")
		twitchURL, exists := e.DOM.Parent().Parent().Find(`a[aria-label="Visit twitch profile"]`).Attr("href")

		if !exists || twitchURL == "" || username == "" {
			return
		}

		parts := strings.Split(twitchURL, "/")
		twitchName := parts[len(parts)-1]

		player := scrapedPlayer{
			Name:   username,
			Tag:    tag,
			Twitch: twitchName,
		}

		logVerbose(cfg.Verbose, "Found player: %s#%s", player.Name, player.Tag)
		scrapedPlayerChan <- player
	})

	url := fmt.Sprintf("https://tracker.gg/valorant/leaderboards/ranked/all/default?platform=pc&region=%s&page=1", cfg.Region)
	if err := c.Visit(url); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var wg sync.WaitGroup

	cfg := loadConfig()

	scrapedPlayerChan := make(chan scrapedPlayer, 100)
	translatedItemsChan := make(chan TranslatedItem, 100)

	fmt.Println("--- Neural Theft Scraper ---")
	fmt.Printf(" > Region: %s\n > Output File: %s\n > %.2f pages per min\n > Verbose: %t\n", cfg.Region, cfg.OutputFile, cfg.Speed, cfg.Verbose)

	go scrapeLeaderboard(cfg, scrapedPlayerChan)

	wg.Wait()
	fmt.Println("\n--- Execution Complete ---")
}
