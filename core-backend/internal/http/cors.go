package http

import (
	nethttp "net/http"
	"slices"
	"strings"
)

type CORSConfig struct {
	AllowedOrigins []string
}

var defaultCORSMethods = []string{
	nethttp.MethodGet,
	nethttp.MethodPost,
	nethttp.MethodPut,
	nethttp.MethodDelete,
	nethttp.MethodOptions,
}

var defaultCORSHeaders = []string{
	"Authorization",
	"Content-Type",
}

func CORSMiddleware(cfg CORSConfig) func(nethttp.Handler) nethttp.Handler {
	allowedOrigins := make([]string, 0, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		allowedOrigins = append(allowedOrigins, trimmed)
	}

	allowedMethods := strings.Join(defaultCORSMethods, ", ")
	allowedHeaders := strings.Join(defaultCORSHeaders, ", ")

	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin != "" {
				w.Header().Add("Vary", "Origin")
				w.Header().Add("Vary", "Access-Control-Request-Method")
				w.Header().Add("Vary", "Access-Control-Request-Headers")
			}

			if !isAllowedOrigin(origin, allowedOrigins) {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

			if r.Method == nethttp.MethodOptions {
				w.WriteHeader(nethttp.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}
	return slices.Contains(allowedOrigins, origin)
}
