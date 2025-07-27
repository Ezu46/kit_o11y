package main

import (
	"embed"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

//go:embed static/*
var content embed.FS

type request struct {
	Method   string `json:"method"`           // GET/POST…
	Path     string `json:"path"`             // "/db_tx,/slow"
	RPS      int    `json:"rps"`              // total RPS
	Duration int    `json:"duration"`         // seconds
}

var inFlight int32

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        2000,
		MaxIdleConnsPerHost: 2000,
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	},
}

func main() {
	http.HandleFunc("/api/start", startHandler)

	// serve UI
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, _ := content.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	log.Println("Load generator UI on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var cfg request
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if cfg.Method == "" {
		cfg.Method = http.MethodGet
	}
	if cfg.RPS <= 0 {
		cfg.RPS = 1
	}
	go generateLoad(cfg)
	w.WriteHeader(http.StatusAccepted)
}

func generateLoad(cfg request) {
	atomic.AddInt32(&inFlight, 1)
	defer atomic.AddInt32(&inFlight, -1)

	paths := strings.Split(cfg.Path, ",")
	if len(paths) == 0 {
		return
	}

	// делим RPS поровну между путями
	rpsPerPath := cfg.RPS / len(paths)
	if rpsPerPath == 0 {
		rpsPerPath = 1
	}

	for _, p := range paths {
		target := "http://app:8000" + strings.TrimSpace(p)
		go fire(target, cfg.Method, rpsPerPath, cfg.Duration)
	}

	log.Printf("▶️  Total %d RPS for %d s across %d path(s): %s",
		cfg.RPS, cfg.Duration, len(paths), cfg.Path)
}

func fire(target, method string, rps, duration int) {
	workers := 8
	rpsPerWorker := rps / workers
	if rpsPerWorker == 0 {
		rpsPerWorker = 1
	}

	end := time.Now().Add(time.Duration(duration) * time.Second)

	for i := 0; i < workers; i++ {
		go func() {
			t := time.NewTicker(time.Second / time.Duration(rpsPerWorker))
			defer t.Stop()

			for now := range t.C {
				if now.After(end) {
					return
				}
				req, _ := http.NewRequest(method, target, nil)
				if resp, err := httpClient.Do(req); err == nil {
					resp.Body.Close()
				}
			}
		}()
	}
}
