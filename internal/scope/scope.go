package scope

import (
	"net"
	"strings"
)

type Rule struct {
	Pattern string
	Action  string
}

type Matcher struct {
	includes []Rule
	excludes []Rule
}

func New(rules []Rule) Matcher {
	m := Matcher{}
	for _, r := range rules {
		if strings.EqualFold(r.Action, "exclude") {
			m.excludes = append(m.excludes, r)
		} else {
			m.includes = append(m.includes, r)
		}
	}
	return m
}

func (m Matcher) Allows(host string) bool {
	host = normalize(host)
	if host == "" {
		return false
	}
	for _, r := range m.excludes {
		if matches(host, r.Pattern) {
			return false
		}
	}
	if len(m.includes) == 0 {
		return false
	}
	for _, r := range m.includes {
		if matches(host, r.Pattern) {
			return true
		}
	}
	return false
}

func normalize(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

func matches(host, pattern string) bool {
	pattern = normalize(pattern)
	if ip := net.ParseIP(host); ip != nil {
		return ip.String() == pattern
	}
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*.")
		return host != suffix && strings.HasSuffix(host, "."+suffix)
	}
	return host == pattern
}
