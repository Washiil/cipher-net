package cmd

import (
	"context"
	"flag"
	"fmt"
	"sync"

	"github.com/washiil/neural-theft/internal/config"
	"github.com/washiil/neural-theft/internal/database"
	"github.com/washiil/neural-theft/internal/processor"
	"github.com/washiil/neural-theft/internal/scraper"
)

func Run() {
	region := flag.String("region", "na", "Leaderboard region (e.g., na, eu, ap)")
	output := flag.String("output", "data.db", "Output SQLite database file")
	speed := flag.Float64("speed", 75, "Scrape speed (users per minute)")
	token := flag.String("token", "", "API token for UUID lookup")
	verbose := flag.Bool("verbose", false, "Enable verbose debug output")
	flag.Parse()

	cfg := config.Config{
		Region:     *region,
		OutputFile: *output,
		Speed:      *speed,
		Token:      *token,
		Verbose:    *verbose,
	}
	cfg = config.ResolveToken(cfg)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rawPlayers := scraper.ScrapePlayers(ctx, cfg, &wg)
	uuidPlayers := processor.AddUUIDs(ctx, cfg, rawPlayers, cfg.Token, &wg)
	database.SaveToDatabase(ctx, cfg, uuidPlayers, &wg)

	fmt.Println("--- Neural Theft Scraper ---")
	fmt.Printf(" > Region: %s\n > Output File: %s\n > %.2f users/min\n > Verbose: %t\n", cfg.Region, cfg.OutputFile, cfg.Speed, cfg.Verbose)

	wg.Wait()
	fmt.Println("\n--- Execution Complete ---")
}
