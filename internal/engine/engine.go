package engine

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"reconflow/internal/pipeline"
	"reconflow/internal/scope"
	"reconflow/internal/storage"
)

type Result struct {
	Hostname   string
	Resolved   bool
	HTTPStatus int
	Error      string
}
type Engine struct {
	Store     *storage.Store
	ProjectID string
	Scope     scope.Matcher
	Client    *http.Client
}

func (e *Engine) Run(ctx context.Context, candidates []string) ([]Result, error) {
	client := e.Client
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	results := make([]Result, 0)
	for _, raw := range pipeline.Unique(candidates) {
		host, err := pipeline.Hostname(raw)
		if err != nil || !e.Scope.Allows(host) {
			continue
		}
		result := Result{Hostname: host}
		now := time.Now().UTC().Format(time.RFC3339)
		ips, dnsErr := net.DefaultResolver.LookupHost(ctx, host)
		status := "unresolved"
		if dnsErr == nil && len(ips) > 0 {
			result.Resolved = true
			status = "resolved"
		}
		assetID, storeErr := e.Store.UpsertAsset(ctx, e.ProjectID, host, status, now)
		if storeErr != nil {
			return nil, storeErr
		}
		if dnsErr != nil {
			_ = e.Store.AddDNS(ctx, assetID, "A", "", dnsErr.Error(), now)
			result.Error = dnsErr.Error()
			results = append(results, result)
			continue
		}
		for _, ip := range ips {
			_ = e.Store.AddDNS(ctx, assetID, "A", ip, "ok", now)
		}
		for _, scheme := range []string{"https", "http"} {
			req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s://%s", scheme, host), nil)
			if reqErr != nil {
				continue
			}
			req.Header.Set("Range", "bytes=0-32767")
			resp, probeErr := client.Do(req)
			if probeErr != nil {
				continue
			}
			bodyBytes, _ := io.CopyN(io.Discard, resp.Body, 32768)
			_ = resp.Body.Close()
			_ = e.Store.AddHTTP(ctx, assetID, scheme, map[bool]int{true: 443, false: 80}[scheme == "https"], resp.StatusCode, int(bodyBytes), now)
			result.HTTPStatus = resp.StatusCode
			break
		}
		results = append(results, result)
	}
	return results, nil
}
