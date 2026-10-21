package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	config, err := read_config("config.json")

	if err != nil {
		log.Fatal(err)
	}

	reg := prometheus.NewRegistry()

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
		Addr:              ":2112",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	client := &http.Client{
		Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(time.Duration(config.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	runRound(ctx, client, config.Entries)
	for {
		select {
		case <-ctx.Done():
			log.Println("Done")
			return
		case <-ticker.C:
			runRound(ctx, client, config.Entries)
		}
	}
}

func runRound(ctx context.Context, client *http.Client, entries []ConfigEntry) {
	results := make(chan HealthcheckResult, len(entries))

	var wg sync.WaitGroup
	for _, e := range entries {
		wg.Go(func() { results <- checkUrl(e, client, ctx) })
	}
	wg.Wait()
	close(results)

	if ctx.Err() != nil {
		return
	}
	for r := range results {
		display(r)
		recordResult(r)
	}
}
