package scraper

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/washiil/cipher-net/internal/config"
	"github.com/washiil/cipher-net/internal/utils"

	"github.com/gocolly/colly"
)

type ScrapedPlayer struct {
	Name   string
	Tag    string
	Twitch string
}

func ScrapePlayers(ctx context.Context, cfg config.Config, wg *sync.WaitGroup) <-chan ScrapedPlayer {
	out := make(chan ScrapedPlayer, int(math.Floor(cfg.Speed)))

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)

		c := colly.NewCollector(
			colly.UserAgent("Mozilla/5.0"),
		)

		c.OnError(func(r *colly.Response, err error) {
			utils.Verbose(cfg.Verbose, "Request to %s failed: %v", r.Request.URL, err)
		})

		for page := 1; ; page++ {
			foundAny := false

			c.OnHTML("td a[href*='/valorant/profile']", func(e *colly.HTMLElement) {
				if ctx.Err() != nil {
					return
				}
				foundAny = true

				username := strings.TrimSuffix(strings.TrimSpace(e.ChildText("span.v3-trnign .max-w-full.truncate")), "#")
				tag := strings.TrimPrefix(strings.TrimSpace(e.ChildText("span.v3-trnign__discriminator")), "#")

				twitchURL, exists := e.DOM.Parent().Parent().
					Find(`a[aria-label="Visit twitch profile"]`).Attr("href")
				if !exists || twitchURL == "" || username == "" {
					return
				}

				parts := strings.Split(twitchURL, "/")
				twitchName := parts[len(parts)-1]

				select {
				case <-ctx.Done():
					return
				case out <- ScrapedPlayer{Name: username, Tag: tag, Twitch: twitchName}:
					utils.Verbose(cfg.Verbose, "Found player: %s#%s", username, tag)
				}
			})

			url := fmt.Sprintf("https://tracker.gg/valorant/leaderboards/ranked/all/default?platform=pc&region=%s&page=%d", cfg.Region, page)
			if err := c.Visit(url); err != nil {
				utils.Verbose(cfg.Verbose, "Visit error on page %d: %v", page, err)
				break
			}

			if !foundAny {
				utils.Verbose(cfg.Verbose, "No players found on page %d", page)
				break
			}
		}
	}()

	return out
}
