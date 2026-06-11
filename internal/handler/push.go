// Package handler provides HTTP endpoints and clients for push/pull mechanisms.
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dlcuy22/exporter-agent/internal/daemon"
)

/*
StartPushWorker starts a background worker that periodically collects and POSTs metrics.

	params:
		ctx: Context for controlling the worker lifecycle
		d: The daemon instance
*/
func StartPushWorker(ctx context.Context, d *daemon.Daemon) {
	url := d.Config().PushURL
	if url == "" {
		return // Push mode disabled
	}

	intervalStr := d.Config().PushInterval
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Printf("Invalid push interval %s, defaulting to 15s. Error: %v", intervalStr, err)
		interval = 15 * time.Second
	}

	log.Printf("Starting push worker to %s every %s", url, interval)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Push worker stopped")
			return
		case <-ticker.C:
			// Perform sequential collection and push
			metrics, err := d.CollectAll(ctx)
			if err != nil {
				log.Printf("Push worker: failed to collect metrics: %v", err)
				continue
			}

			payload, err := json.Marshal(metrics)
			if err != nil {
				log.Printf("Push worker: failed to serialize payload: %v", err)
				continue
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
			if err != nil {
				log.Printf("Push worker: failed to create request: %v", err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Push worker: request failed: %v", err)
				continue
			}
			
			// We only care about the status code
			resp.Body.Close()
			if resp.StatusCode >= 400 {
				log.Printf("Push worker: unexpected status code %d", resp.StatusCode)
			}
		}
	}
}
