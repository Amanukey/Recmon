package main

import (
	"crypto/rand"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

//go:embed web/*
var webFS embed.FS

func main() {
	token := make([]byte, 18)
	if _, err := rand.Read(token); err != nil {
		log.Fatal(err)
	}
	accessToken := fmt.Sprintf("%x", token)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"status": "ok", "service": "reconflow"})
	})
	mux.HandleFunc("/api/projects", projectsHandler)
	mux.HandleFunc("/api/scans", scansHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := webFS.ReadFile("web/index.html")
		if err != nil {
			http.Error(w, "dashboard unavailable", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	handler := withSecurity(mux, accessToken)
	url := "http://" + listener.Addr().String() + "/?token=" + accessToken
	fmt.Fprintln(os.Stdout, "ReconFlow is running")
	fmt.Fprintln(os.Stdout, "Local access:", url)
	log.Fatal(http.Serve(listener, handler))
}

func withSecurity(next http.Handler, token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if !strings.HasPrefix(host, "127.0.0.1:") && host != "127.0.0.1" && !strings.HasPrefix(host, "localhost:") && host != "localhost" {
			http.Error(w, "invalid host", http.StatusForbidden)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/api/health" && r.URL.Query().Get("token") != token && r.Header.Get("X-ReconFlow-Token") != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
