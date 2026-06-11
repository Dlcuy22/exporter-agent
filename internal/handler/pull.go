// Package handler provides HTTP endpoints and clients for push/pull mechanisms.
//
// Key Components:
//   - StartPullServer: Starts an HTTP server for GET /metrics
//
// Dependencies:
//   - net/http
//   - encoding/json
//   - github.com/dlcuy22/exporter-agent/internal/daemon
//
// Error Types:
//   - None
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dlcuy22/exporter-agent/internal/daemon"
)

/*
StartPullServer initializes and runs the HTTP server for the pull endpoint.

	params:
		d: The daemon instance
	returns:
		*http.Server: The initialized and running HTTP server
*/
func StartPullServer(d *daemon.Daemon) *http.Server {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if d.Config().PullAuth {
			token := r.Header.Get("X-Token")
			if token == "" || token != d.Config().PullToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		metrics, err := d.GetLatestMetrics(ctx)
		if err != nil {
			http.Error(w, "Failed to collect metrics", http.StatusInternalServerError)
			log.Printf("Pull handler error: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			log.Printf("Pull handler JSON encode error: %v", err)
		}
	})

	port := d.Config().PullPort
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		log.Printf("Starting pull server on :%d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Pull server failed: %v", err)
		}
	}()

	return srv
}
