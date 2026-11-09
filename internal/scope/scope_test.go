package scope

import "testing"

func TestMatcherKeepsWildcardBoundaries(t *testing.T) {
	m := New([]Rule{{Pattern: "*.example.com", Action: "include"}})
	for host, want := range map[string]bool{"api.example.com": true, "example.com": false, "api.example.com.evil.net": false} {
		if got := m.Allows(host); got != want {
			t.Fatalf("Allows(%q)=%v, want %v", host, got, want)
		}
	}
}

func TestExclusionsWin(t *testing.T) {
	m := New([]Rule{{Pattern: "*.example.com", Action: "include"}, {Pattern: "admin.example.com", Action: "exclude"}})
	if m.Allows("admin.example.com") {
		t.Fatal("excluded host was allowed")
	}
}
