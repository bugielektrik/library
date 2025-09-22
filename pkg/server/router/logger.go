package router

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

// LoggerWithSkips returns a middleware that disables request logging for
// matching paths. It deliberately uses chi's middleware.Logger for logging
// all non-skipped requests. Do NOT add r.Use(middleware.Logger) elsewhere.
//
// Patterns supported:
// - Exact paths: "/health"
// - Prefix wildcard: "/swagger/{*}"  -> matches any path whose first segment is "swagger"
// - Parameterized paths: "/billings/{id}/status" -> matches any path with same segment count and same first/last segments
func LoggerWithSkips(patterns []string) func(http.Handler) http.Handler {
	m := newPathMatcher(patterns)

	return func(next http.Handler) http.Handler {
		// chi's logger handler; we delegate to it when we want logging.
		chiLogger := middleware.Logger(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := normalizePath(r.URL.Path)

			// If path is configured to be skipped, call next without logging.
			if m.shouldSkip(path) {
				next.ServeHTTP(w, r)
				return
			}

			// Otherwise use chi's logger which will log and then call next.
			chiLogger.ServeHTTP(w, r)
		})
	}
}

// pathMatcher holds precomputed sets for fast matching.
type pathMatcher struct {
	exact  map[string]struct{} // exact matches: "/health"
	prefix map[string]struct{} // prefix matches by first segment: "/swagger"
	dyn    map[string]struct{} // dynamic matches keyed by "len|first|last"
}

// newPathMatcher builds the matcher from user-provided patterns.
func newPathMatcher(patterns []string) *pathMatcher {
	pm := &pathMatcher{
		exact:  make(map[string]struct{}),
		prefix: make(map[string]struct{}),
		dyn:    make(map[string]struct{}),
	}

	for _, raw := range patterns {
		p := normalizePath(strings.TrimSpace(raw))
		if p == "" {
			continue
		}

		// prefix wildcard: ends with "{*}"
		if strings.HasSuffix(p, "{*}") {
			base := strings.TrimSuffix(p, "{*}")
			// use first segment to determine prefix
			seg := firstSegment(base)
			if seg != "" {
				pm.prefix["/"+seg] = struct{}{}
			}
			continue
		}

		// parameterized path contains '{' (simple heuristic)
		if strings.Contains(p, "{") {
			segs := splitSegments(p)
			if len(segs) == 0 {
				continue
			}
			// key: length|first|last, where pattern params are normalized to "*"
			key := buildDynKeyFromSegments(segs)
			pm.dyn[key] = struct{}{}
			continue
		}

		// exact path
		pm.exact[p] = struct{}{}
	}

	return pm
}

// shouldSkip returns true if path must bypass logging.
func (pm *pathMatcher) shouldSkip(path string) bool {
	// exact match
	if _, ok := pm.exact[path]; ok {
		return true
	}

	// prefix by first segment
	if _, ok := pm.prefix["/"+firstSegment(path)]; ok {
		return true
	}

	// dynamic match: compare length, first and last segments
	segs := splitSegments(path)
	key := strconv.Itoa(len(segs)) + "|" + segs[0] + "|" + segs[len(segs)-1]
	if _, ok := pm.dyn[key]; ok {
		return true
	}

	return false
}

// Helpers

// normalizePath lowercases the path and removes a trailing slash (except for "/").
func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	p = strings.ToLower(p)
	if len(p) > 1 && p[len(p)-1] == '/' {
		return p[:len(p)-1]
	}
	return p
}

// splitSegments splits a path into its segments. For root ("" or "/") it returns ["/"].
func splitSegments(p string) []string {
	if p == "" || p == "/" {
		return []string{"/"}
	}
	return strings.Split(strings.TrimPrefix(p, "/"), "/")
}

// firstSegment returns the first segment of the path (or "" if none).
func firstSegment(p string) string {
	segs := splitSegments(p)
	if len(segs) == 0 {
		return ""
	}
	return segs[0]
}

// buildDynKeyFromSegments constructs the dyn map key from pattern segments.
// It replaces parameter segments like "{id}" with "*" to normalize keys.
func buildDynKeyFromSegments(segs []string) string {
	n := len(segs)
	first := normalizePatternSegment(segs[0])
	last := normalizePatternSegment(segs[n-1])
	return strconv.Itoa(n) + "|" + first + "|" + last
}

// normalizePatternSegment turns parameter segments "{...}" into "*" else returns as-is.
func normalizePatternSegment(seg string) string {
	if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
		return "*"
	}
	return seg
}
