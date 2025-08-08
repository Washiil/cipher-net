package config

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Region     string
	OutputFile string
	Speed      float64
	Token      string
	Verbose    bool
}

func ResolveToken(cfg Config) Config {
	if cfg.Token != "" {
		return cfg
	}
	log.Println("No token specified, attempting to read from .env...")
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env")
	}
	envToken := os.Getenv("API_KEY")
	if envToken == "" {
		log.Fatal("API_KEY not set in .env")
	}
	cfg.Token = envToken
	return cfg
}

func Load() Config {
	region := flag.String("region", "na", "Leaderboard to scrape")
	outputFile := flag.String("output", "data.db", "Output file")
	speed := flag.Float64("speed", 75, "Users per minute")
	token := flag.String("token", "", "API Token")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	if *token == "" {
		log.Println("No token specified, loading .env...")
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env")
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
