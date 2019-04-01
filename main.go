package main

import (
	"fetcherc/fetcher"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	f := fetcher.New(2*time.Second, 5*time.Second)
	quit := f.Run(getFetch())

	// Following are debounced
	go func() { f.TriggerManualFetch() }()
	go func() { f.TriggerManualFetch() }()
	go func() { f.TriggerManualFetch() }()

	time.Sleep(6 * time.Second)
	quit <- true
	payload := f.GetPayload()
	fmt.Println(payload)
}

func getFetch() fetcher.Fetch {
	return func(res chan interface{}, quit chan bool) {
		i := rand.Intn(3)
		seconds, _ := time.ParseDuration(fmt.Sprintf("%ds", i))
		select {
		case <-time.After(seconds):
			fmt.Println("Fetch complete")
			fmt.Println()
			res <- fmt.Sprintf("foo")
			return
		case <-quit:
			fmt.Println("exiting the fetch and swallowing the payload")
			return
		}
	}
}
