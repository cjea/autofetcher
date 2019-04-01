package main

import (
	"fetcherc/fetcher"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	f := fetcher.New(2*time.Second, 5*time.Second)
	go func() { f.Run(getFetch()) }()

	time.Sleep(2 * time.Second)

	// Following are debounced
	go func() { f.TriggerManualFetch() }()
	go func() { f.TriggerManualFetch() }()
	go func() { f.TriggerManualFetch() }()

	time.Sleep(6 * time.Second)
	fmt.Println(f.Payload)
}

func getFetch() fetcher.Fetch {
	count := 1
	return func(res chan interface{}, quit chan bool) {
		i := rand.Intn(3)
		seconds, _ := time.ParseDuration(fmt.Sprintf("%ds", i))
		// fmt.Println(fmt.Sprintf("Fetching... This will take %d seconds", i))
		select {
		case <-time.After(seconds):
			fmt.Println("Fetch complete\n")
			res <- fmt.Sprintf("foo %d", count)
			count = count + 1
		case <-quit:
			fmt.Println("exiting the fetch and swallowing the payload")
			return
		}
	}
}
