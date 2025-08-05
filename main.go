package main

import (
	"flag"
	"fmt"
)

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

	fmt.Println("\n--- Execution Complete ---")
}
