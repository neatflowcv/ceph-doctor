package domain

import (
	"errors"
	"strings"
)

var (
	ErrEmptyHosts    = errors.New("cluster hosts are empty")
	ErrEmptyHost     = errors.New("cluster host is empty")
	ErrDuplicateHost = errors.New("cluster host is duplicated")
)

type Hosts struct {
	values []string
}

func NewHosts(hosts []string) (*Hosts, error) {
	if len(hosts) == 0 {
		return nil, ErrEmptyHosts
	}

	normalizedHosts := make([]string, 0, len(hosts))
	seen := make(map[string]struct{}, len(hosts))

	for _, host := range hosts {
		if strings.TrimSpace(host) == "" {
			return nil, ErrEmptyHost
		}

		normalizedHost := normalizeHost(host)
		if _, ok := seen[normalizedHost]; ok {
			return nil, ErrDuplicateHost
		}

		seen[normalizedHost] = struct{}{}
		normalizedHosts = append(normalizedHosts, normalizedHost)
	}

	return &Hosts{values: normalizedHosts}, nil
}

func (h *Hosts) Values() []string {
	return append([]string(nil), h.values...)
}

func normalizeHost(host string) string {
	for i := range len(host) {
		if host[i] == ':' {
			return host
		}
	}

	return host + ":3300"
}
