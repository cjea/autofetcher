package fetcher

import (
	"fmt"
	"sync"
	"time"
)

// Fetcher is the main type
type Fetcher struct {
	Payload            interface{}
	AutoFetchCadence   time.Duration
	DebouncePeriod     time.Duration
	FetchTimeout       time.Duration
	LastManualFetch    time.Time
	manualFetchRequest chan bool
	nextPayload        chan interface{}
	Mux                sync.Mutex
}

// Fetch is the function signature for a fetch operation.
// The result of the fetch should be written to the out channel
type Fetch = func(out chan interface{}, quit chan bool)

// New returns a new fetcher
func New(autoFetchCadence, debouncePeriod time.Duration) *Fetcher {
	manualFetchRequest := make(chan bool)
	next := make(chan interface{})
	return &Fetcher{
		manualFetchRequest: manualFetchRequest,
		Payload:            nil,
		AutoFetchCadence:   autoFetchCadence,
		DebouncePeriod:     debouncePeriod,
		FetchTimeout:       10 * time.Second,
		nextPayload:        next,
	}
}

// LoopSetPayload watches to set payload
func (f *Fetcher) LoopSetPayload() {
	go func() {
		for {
			f.Payload = <-f.nextPayload
		}
	}()
}

// WaitFor either gets the payload or cancels the fetch
func (f *Fetcher) WaitFor(fn Fetch) {
	res := make(chan interface{})
	quit := make(chan bool)
	go func() { fn(res, quit) }()
	select {
	case next := <-res:
		f.nextPayload <- next
		return
	case <-time.After(f.FetchTimeout):
		quit <- true
		fmt.Println("fetch timed out")
		return
	}
}

// LoopAutoFetch loops and fetches automatically
func (f *Fetcher) LoopAutoFetch(fn Fetch) {
	go func() {
		for {
			fmt.Println("Beginning auto fetch")
			logTime()
			f.WaitFor(fn)
		}
	}()
}

// TriggerManualFetch will cause a manual fetch to be performed
// in a concurrent-safe way
func (f *Fetcher) TriggerManualFetch() {
	f.manualFetchRequest <- true
}

// ManualFetchDebounce loops and fetches automatically
func (f *Fetcher) ManualFetchDebounce(fn Fetch) {
	go func() {
		for {
			<-f.manualFetchRequest
			fmt.Println("Beginning manual fetch")
			logTime()
			f.WaitFor(fn)
			f.LastManualFetch = time.Now().UTC()
		}
	}()
}

// Run runs the fetching behavior.
// 	- run the function at least once every AutoFetchCadence time period
// 	- run manual invocations at most once every DebouncePeriod time period
func (f *Fetcher) Run(fn Fetch) chan bool {
	quit := make(chan bool)
	f.LoopSetPayload()
	go func() {
		for {
			select {
			case <-time.After(f.AutoFetchCadence):
				fmt.Println("Beginning auto fetch")
				logTime()
				f.WaitFor(fn)
			case <-f.manualFetchRequest:
				if time.Since(f.LastManualFetch) >= f.DebouncePeriod {
					fmt.Println("Beginning manual fetch")
					logTime()
					f.WaitFor(fn)
					f.LastManualFetch = time.Now().UTC()
				}
			case <-quit:
				fmt.Println("returning!")
				return
			}
		}
	}()
	return quit
	// f.LoopAutoFetch(fn)
	// f.ManualFetchDebounce(fn)
}

func logTime() {
	_, m, s := time.Now().Clock()
	fmt.Println(fmt.Sprintf("Clock:\t%dm%ds", m, s))
	fmt.Println()
}
