package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/washiil/cipher-net/internal/config"
	"github.com/washiil/cipher-net/internal/processor"
)

func SaveToDatabase(ctx context.Context, cfg config.Config, in <-chan processor.DatabasePlayer, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		db, err := sql.Open("sqlite", cfg.OutputFile)
		if err != nil {
			log.Fatalf("Error opening DB: %v", err)
		}
		defer db.Close()

		schema := `CREATE TABLE IF NOT EXISTS players (
			uuid TEXT PRIMARY KEY,
			name TEXT,
			tag TEXT,
			twitch TEXT
		);`
		if _, err := db.ExecContext(ctx, schema); err != nil {
			log.Fatalf("Schema error: %v", err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Fatalf("Transaction error: %v", err)
		}

		insertCount := 0
		for p := range in {
			if ctx.Err() != nil {
				_ = tx.Rollback()
				return
			}

			_, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO players (uuid, name, tag, twitch) VALUES (?, ?, ?, ?)", p.UUID, p.Name, p.Tag, p.Twitch)
			if err != nil {
				log.Printf("Insert error for %s: %v", p.UUID, err)
				continue
			}

			insertCount++
			if insertCount%10 == 0 {
				if err := tx.Commit(); err != nil {
					log.Fatalf("Commit error: %v", err)
				}
				fmt.Printf("\r > Saved %d players...", insertCount)
				tx, err = db.BeginTx(ctx, nil)
				if err != nil {
					log.Fatalf("Transaction restart error: %v", err)
				}
			}
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("Final commit error: %v", err)
		}
		fmt.Printf("\n > Finished saving %d players.\n", insertCount)
	}()
}
