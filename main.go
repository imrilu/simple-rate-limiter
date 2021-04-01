package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type RequestsQueue struct {
	queue []time.Time
	lock  sync.Mutex
}

type RateLimiter struct {
	ttl       time.Duration
	threshold int
	requests  map[string]*RequestsQueue
	lock      sync.Mutex
}

type ReportBody struct {
	Url string
}

func (rl *RateLimiter) parseBody(w http.ResponseWriter, r *http.Request) string {
	decoder := json.NewDecoder(r.Body)
	var reportBody ReportBody
	err := decoder.Decode(&reportBody)
	if err != nil {
		panic(err)
	}
	log.Println("func parseBody with body recieved: ", reportBody)
	return reportBody.Url
}

func (rl *RateLimiter) serve(url string) bool {
	log.Println("func serve with url:", url)
	// locking map structure for value search/insert
	rl.lock.Lock()
	requestsQueue, exist := rl.requests[url]
	if !exist {
		rl.requests[url] = NewRequestsQueue()
		requestsQueue = rl.requests[url]
	}
	rl.lock.Unlock()
	log.Println("url:", url, "\n Queue before serve: ", requestsQueue.queue)

	requestsQueue.lock.Lock()
	// clearing timeout requests from queue
	now := time.Now()
	start := 0
	for i := range requestsQueue.queue {
		if now.Sub(requestsQueue.queue[i]) > rl.ttl {
			start += 1
			log.Println("url:", url, "\nRemoved element:", i)
		}
	}
	requestsQueue.queue = requestsQueue.queue[start:]
	// comparing threshold with queue size
	if len(requestsQueue.queue) < rl.threshold {
		requestsQueue.queue = append(requestsQueue.queue, now)
		requestsQueue.lock.Unlock()
		return true
	}
	requestsQueue.lock.Unlock()
	return false
}

func (rl *RateLimiter) report(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" && r.Header.Get("content-type") == "application/json" {
		url := rl.parseBody(w, r)
		response := NewResponse(rl.serve(url))
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
	}
}

func NewRateLimiter(threshold int, ttl int) *RateLimiter {
	return &RateLimiter{
		ttl:       time.Duration(ttl) * time.Millisecond,
		threshold: threshold,
		requests:  make(map[string]*RequestsQueue),
	}
}

func NewRequestsQueue() *RequestsQueue {
	return &RequestsQueue{}
}

func NewResponse(served bool) map[string]bool {
	response := make(map[string]bool)
	response["block"] = served
	return response
}

func main() {
	// initing threshold and ttl variables from command line args
	threshold, err1 := strconv.Atoi(os.Args[1])
	ttl, err2 := strconv.Atoi(os.Args[2])
	if err1 != nil || err2 != nil {
		log.Fatal("error parsing command line args")
		return
	}
	RateLimiter := NewRateLimiter(threshold, ttl)
	http.HandleFunc("/report", RateLimiter.report)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
