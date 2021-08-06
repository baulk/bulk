package netutils

import (
	"errors"
	"os"
	"strings"
)

// error
var (
	ErrProxyNotConfigured = errors.New("Proxy is not configured correctly")
)

// ProxySettings todo
type ProxySettings struct {
	ProxyServer   string
	ProxyOverride string // aka no proxy
	sep           string
}

func getEnvAny(names ...string) string {
	for _, n := range names {
		if val := os.Getenv(n); val != "" {
			return val
		}
	}
	return ""
}

func canonicalProxyURL(s string) string {
	if !strings.Contains(s, "://") {
		return "http://" + s // avoid proxy url parse failed
	}
	return s
}
