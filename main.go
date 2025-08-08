package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
	_ "modernc.org/sqlite"
)

type Config struct {
	Region     string
	OutputFile string
	Speed      float64
	Token      string
	Verbose    bool
}

type scrapedPlayer struct {
	Name   string
	Tag    string
	Twitch string
}

type databasePlayer struct {
	Name   string
	Tag    string
	Twitch string
	UUID   string
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

func scrapePlayers(ctx context.Context, cfg Config, wg *sync.WaitGroup) <-chan scrapedPlayer {
	out := make(chan scrapedPlayer, int(math.Floor(cfg.Speed)))

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)

		c := colly.NewCollector(
			colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
		)

		// c.Limit(&colly.LimitRule{
		// 	DomainGlob:  "*tracker.gg*",
		// 	Delay:       time.Minute / time.Duration(cfg.Speed),
		// 	Parallelism: 1,
		// })

		c.OnError(func(r *colly.Response, err error) {
			logVerbose(cfg.Verbose, "Request to %s failed: %v", r.Request.URL, err) // ADDED: error logging
		})

		for page := 1; ; page++ {
			foundAny := false

			c.OnHTML("td a[href*='/valorant/profile']", func(e *colly.HTMLElement) {
				if ctx.Err() != nil {
					return
				}

				foundAny = true

				rawName := e.ChildText("span.v3-trnign .max-w-full.truncate")
				username := strings.TrimSuffix(strings.TrimSpace(rawName), "#")
				rawTag := e.ChildText("span.v3-trnign__discriminator")
				tag := strings.TrimPrefix(strings.TrimSpace(rawTag), "#")

				twitchURL, exists := e.DOM.Parent().Parent().
					Find(`a[aria-label="Visit twitch profile"]`).Attr("href")
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

				select {
				case <-ctx.Done():
					return
				case out <- player:
				}
			})

			// build URL and visit
			url := fmt.Sprintf(
				"https://tracker.gg/valorant/leaderboards/ranked/all/default?platform=pc&region=%s&page=%d",
				cfg.Region, page,
			)
			if err := c.Visit(url); err != nil {
				logVerbose(cfg.Verbose, "Visit error on page %d: %v", page, err)
				break
			}

			if ctx.Err() != nil {
				break
			}
			// if no players found on this page, we’re past the last page
			if !foundAny {
				logVerbose(cfg.Verbose, "No players on page %d—stopping", page)
				break
			}
		}
	}()

	return out
}

func addUUIDs(ctx context.Context, cfg Config, in <-chan scrapedPlayer, token string, wg *sync.WaitGroup) <-chan databasePlayer {
	out := make(chan databasePlayer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)

		limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(cfg.Speed)), 1)

		for player := range in {
			if err := limiter.Wait(ctx); err != nil {
				return
			}

			uuid, err := NameToUUID(player.Name, player.Tag, token)
			if err != nil {
				logVerbose(true, "UUID lookup failed for %s#%s: %v", player.Name, player.Tag, err) // ADDED: log lookup errors
				continue
			}

			select {
			case <-ctx.Done():
				return
			case out <- databasePlayer{
				Name:   player.Name,
				Tag:    player.Tag,
				Twitch: player.Twitch,
				UUID:   uuid,
			}:
			}
		}
	}()

	return out
}

func saveToDatabase(ctx context.Context, cfg Config, player <-chan databasePlayer, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		db, err := sql.Open("sqlite", cfg.OutputFile)
		if err != nil {
			log.Fatalf("Error opening database: %v", err)
		}
		defer db.Close()

		createTableSQL := `CREATE TABLE IF NOT EXISTS players (
			"uuid" TEXT NOT NULL PRIMARY KEY,
			"name" TEXT,
			"tag" TEXT,
			"twitch" TEXT
		);`
		if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
			log.Fatalf("Error creating table: %v", err)
		}

		// Begin the first transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Fatalf("Error beginning transaction: %v", err)
		}

		insertCount := 0
		for item := range player {
			// Check for cancellation before processing the item
			if ctx.Err() != nil {
				log.Println("Database saver context cancelled, rolling back transaction.")
				_ = tx.Rollback()
				return
			}

			query := "INSERT OR IGNORE INTO players (uuid, name, tag, twitch) VALUES (?, ?, ?, ?)"
			if _, err := tx.ExecContext(ctx, query, item.UUID, item.Name, item.Tag, item.Twitch); err != nil {
				log.Printf("Error inserting player %s: %v", item.UUID, err)
				continue
			}

			insertCount++

			if insertCount%10 == 0 {
				if err := tx.Commit(); err != nil {
					log.Fatalf("Error committing transaction: %v", err)
				}
				fmt.Printf("\r > Saved %d players...", insertCount)
				tx, err = db.BeginTx(ctx, nil)
				if err != nil {
					log.Fatalf("Error beginning new transaction: %v", err)
				}
			}
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("Error committing final transaction: %v", err)
		}

		fmt.Printf("\n > Finalizing database... Saved a total of %d players.\n", insertCount)
	}()
}

func main() {
	var wg sync.WaitGroup

	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rawPlayerChan := scrapePlayers(ctx, cfg, &wg)
	databaseReadyPlayer := addUUIDs(ctx, cfg, rawPlayerChan, cfg.Token, &wg)
	saveToDatabase(ctx, cfg, databaseReadyPlayer, &wg)

	fmt.Println("--- Neural Theft Scraper ---")
	fmt.Printf(" > Region: %s\n > Output File: %s\n > %.2f pages per min\n > Verbose: %t\n", cfg.Region, cfg.OutputFile, cfg.Speed, cfg.Verbose)

	wg.Wait()
	fmt.Println("\n--- Execution Complete ---")
}
