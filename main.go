package main

import (
	"fetcherc/fetcher"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("Initializing...")
	f := fetcher.New(1*time.Second, 5*time.Second)
	f.Run(getFetch())
	select {
	case <-time.After(10 * time.Second):
		fmt.Println(f.Payload)
	}
}

func getFetch() fetcher.Fetch {
	count := 0
	return func(res chan interface{}, quit chan bool) {
		i := rand.Intn(3)
		// i := 0.5
		seconds, _ := time.ParseDuration(fmt.Sprintf("%ds", i))
		fmt.Println(fmt.Sprintf("Fetching... This will take %d seconds", i))
		// fmt.Println("Fetching... This will take 0.5 seconds")
		select {
		case <-time.After(seconds):
			fmt.Println("Fetch complete")
			res <- fmt.Sprintf("foo %d", count)
			count = count + 1
		case <-quit:
			fmt.Println("exiting the fetch and swallowing the payload")
			return
		}
	}
}
