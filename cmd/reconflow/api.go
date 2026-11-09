package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"reconflow/internal/engine"
	"reconflow/internal/scope"
	"reconflow/internal/storage"
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
		startedAt := time.Now().UTC()
		scanID := "scan_" + strconv.FormatInt(startedAt.UnixNano(), 10)
		writeJSON(w, map[string]any{"id": scanID, "status": "running", "startedAt": startedAt})
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, map[string]any{"id": "scan_local", "status": "running", "sources": []any{map[string]any{"name": "Subfinder", "status": "running"}, map[string]any{"name": "crt.sh", "status": "completed"}, map[string]any{"name": "Assetfinder", "status": "queued"}}, "metrics": map[string]any{"candidates": 128, "validated": 18, "http": 11, "queue": 42}})
}

func engineRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		ProjectID  string       `json:"projectId"`
		Candidates []string     `json:"candidates"`
		Scope      []scope.Rule `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.ProjectID == "" {
		http.Error(w, "projectId and candidates are required", http.StatusBadRequest)
		return
	}
	store, err := storage.Open(":memory:")
	if err != nil {
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	defer store.Close()
	results, err := (&engine.Engine{Store: store, ProjectID: request.ProjectID, Scope: scope.New(request.Scope)}).Run(context.Background(), request.Candidates)
	if err != nil {
		http.Error(w, "engine run failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"results": results, "count": len(results)})
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
		sequenceText := strconv.Itoa(sequence)
		payload, _ := json.Marshal(map[string]any{"message": "Live update " + sequenceText + " received", "sequence": sequence})
		_, _ = w.Write([]byte("id: " + sequenceText + "\nevent: metric\ndata: " + string(payload) + "\n\n"))
		flusher.Flush()
		select {
		case <-time.After(2 * time.Second):
		case <-r.Context().Done():
			return
		}
	}
}
