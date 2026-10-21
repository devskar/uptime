package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HealthcheckResult struct {
	URL           string
	CheckedAt     time.Time
	StatusCode    int
	Status        string
	Duration      time.Duration
	ResponseBytes int64
	ContentType   string
	Up            bool
	Err           error
}

func checkUrl(entry ConfigEntry, client *http.Client, ctx context.Context) (result HealthcheckResult) {
	start := time.Now()
	result.URL = entry.Url

	defer func() {
		result.CheckedAt = start
		result.Duration = time.Since(start)
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, entry.Url, http.NoBody)

	if err != nil {
		result.Err = err
		return result
	}

	resp, err := client.Do(req)

	if err != nil {
		result.Err = err
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Status = resp.Status
	result.ContentType = resp.Header.Get("Content-Type")

	// Drain the body to measure its size and allow connection reuse.
	n, err := io.Copy(io.Discard, resp.Body)
	result.ResponseBytes = n
	if err != nil {
		result.Err = fmt.Errorf("reading body: %w", err)
		return result
	}

	result.Up = resp.StatusCode >= 200 && resp.StatusCode < 300
	return result
}
