package fetcher

import (
	"fmt"
	"time"
)

// Status represents the state of a fetch
type Status int

const (
	// Done indicates that the system is ready to fetch
	Done Status = iota
)

// Fetcher is the main type
type Fetcher struct {
	manualFetchRequest chan bool
	Payload            interface{}
	AutoFetchCadence   time.Duration
	DebouncePeriod     time.Duration
	FetchTimeout       time.Duration
	nextPayload        chan interface{}
}

// Fetch is the function signature for a fetcher.
// The result will be set as Payload field
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
			f.WaitFor(fn)
			time.Sleep(f.AutoFetchCadence)
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
			f.WaitFor(fn)
			time.Sleep(f.DebouncePeriod)
		}
	}()
}

// Run runs the fetching behavior.
// 	- run the function at least once every AutoFetchCadence time period
// 	- run the function at most once every DebouncePeriod time period
func (f *Fetcher) Run(fn Fetch) {
	f.LoopSetPayload()
	f.LoopAutoFetch(fn)
	f.ManualFetchDebounce(fn)
}
