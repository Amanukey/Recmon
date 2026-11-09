package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		writeJSON(w, map[string]any{"id": "proj_local", "name": "Default project", "createdAt": time.Now().UTC()})
		return
	}
	writeJSON(w, []map[string]any{{"id": "proj_local", "name": "Default project", "scope": "*.example.com", "assets": 24, "validated": 18}})
}

func scansHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/events") {
		streamScanEvents(w, r)
		return
	}

	if r.Method == http.MethodPost {
		writeJSON(w, map[string]any{"id": "scan_local", "status": "running", "startedAt": time.Now().UTC()})
		return
	}

	writeJSON(w, map[string]any{"id": "scan_local", "status": "running", "sources": []any{map[string]any{"name": "Subfinder", "status": "running"}, map[string]any{"name": "crt.sh", "status": "completed"}, map[string]any{"name": "Assetfinder", "status": "queued"}}, "metrics": map[string]any{"candidates": 128, "validated": 18, "http": 11, "queue": 42}})
}

func streamScanEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unavailable", http.StatusInternalServerError)
		return
	}

	for sequence := 1; sequence <= 5; sequence++ {
		payload, _ := json.Marshal(map[string]any{"message": "Live update " + string(rune('0'+sequence)) + " received", "sequence": sequence})
		_, _ = w.Write([]byte("id: " + string(rune('0'+sequence)) + "\nevent: metric\ndata: " + string(payload) + "\n\n"))
		flusher.Flush()
		select {
		case <-time.After(2 * time.Second):
		case <-r.Context().Done():
			return
		}
	}
}
