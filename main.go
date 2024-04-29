package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/semaphore"
)

func main() {
	start := time.Now() // record the start time

	urls := []string{
		"https://www.google.com",
		"https://www.github.com",
		"https://badsite.heroku.com",
		"https://www.google.com/21",
	}

	// use a semaphore to limit the number of concurrent requests
	maxConcurrency := int64(10)
	sem := semaphore.NewWeighted(maxConcurrency)
	ctx := context.Background()

	results := make(chan string, len(urls))

	for _, url := range urls {
		// acquire a semaphore weight of 1 before starting a goroutine
		if err := sem.Acquire(ctx, 1); err != nil {
			fmt.Printf("Failed to acquire semaphore: %v\n", err)
			continue
		}

		go checkSite(sem, url, results)
	}

	// wait for all goroutines to release their semaphore weights
	if err := sem.Acquire(ctx, maxConcurrency); err != nil {
		fmt.Printf("Failed to acquire semaphore while waiting: %v\n", err)
	}
	// print results
	for i := 0; i < len(urls); i++ {
		fmt.Println(<-results)
	}

	fmt.Printf("- total time: %s\n", time.Since(start)) // print the elapsed time
}

func checkSite(sem *semaphore.Weighted, url string, results chan<- string) {
	start := time.Now() // record the start time

	defer sem.Release(1)

	res, err := http.Get(url)
	if err != nil {
		results <- fmt.Sprintf("❌ (request failed) %s - %s", url, time.Since(start))
		return
	}

	defer res.Body.Close()

	var emoji string
	if res.StatusCode == http.StatusOK {
		emoji = "✅"
	} else {
		emoji = "❌"
	}
	results <- fmt.Sprintf("%s (%v) %s - %s", emoji, res.StatusCode, url, time.Since(start))
}
