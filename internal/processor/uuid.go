package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/washiil/neural-theft/internal/config"
	"github.com/washiil/neural-theft/internal/scraper"
	"github.com/washiil/neural-theft/internal/utils"

	"golang.org/x/time/rate"
)

type DatabasePlayer struct {
	Name   string
	Tag    string
	Twitch string
	UUID   string
}

func AddUUIDs(ctx context.Context, cfg config.Config, in <-chan scraper.ScrapedPlayer, token string, wg *sync.WaitGroup) <-chan DatabasePlayer {
	out := make(chan DatabasePlayer)
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(out)

		limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(cfg.Speed)), 1)

		for player := range in {
			if err := limiter.Wait(ctx); err != nil {
				return
			}

			uuid, err := nameToUUID(player.Name, player.Tag, token)
			if err != nil {
				utils.Verbose(true, "UUID lookup failed for %s#%s: %v", player.Name, player.Tag, err)
				continue
			}

			select {
			case <-ctx.Done():
				return
			case out <- DatabasePlayer{Name: player.Name, Tag: player.Tag, Twitch: player.Twitch, UUID: uuid}:
			}
		}
	}()
	return out
}

func nameToUUID(name, tag, token string) (string, error) {
	url := fmt.Sprintf("https://api.henrikdev.xyz/valorant/v1/account/%s/%s", name, tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s (status %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Data struct {
			PUUID string `json:"puuid"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Data.PUUID, nil
}
