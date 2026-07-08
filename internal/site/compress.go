package site

import (
	"net/http"
	"strconv"
	"strings"
)

func headCompressionMetadata(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead || r.Header.Get("Range") != "" || !headAcceptsGzip(r.Header.Get("Accept-Encoding")) {
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(&headCompressionResponseWriter{ResponseWriter: w}, r)
	})
}

type headCompressionResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *headCompressionResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	header := w.Header()
	// Ohm marks compressible HEAD responses with Vary, but no body is written to
	// trigger Content-Encoding.
	if headCompressionStatusAllowsBody(status) &&
		header.Get("Content-Encoding") == "" &&
		varyContains(header, "Accept-Encoding") {
		header.Set("Content-Encoding", "gzip")
		header.Del("Content-Length")
		header.Del("Accept-Ranges")
	}

	w.ResponseWriter.WriteHeader(status)
}

func (w *headCompressionResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func headCompressionStatusAllowsBody(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	default:
		return true
	}
}

func headAcceptsGzip(header string) bool {
	var wildcardAccepted bool
	var gzipSeen bool
	var gzipAccepted bool
	for _, value := range strings.Split(header, ",") {
		encoding, quality := parseAcceptEncoding(value)
		switch encoding {
		case "gzip":
			gzipSeen = true
			if quality > 0 {
				gzipAccepted = true
			}
		case "*":
			if quality > 0 {
				wildcardAccepted = true
			}
		}
	}
	if gzipSeen {
		return gzipAccepted
	}
	return wildcardAccepted
}

func parseAcceptEncoding(value string) (string, float64) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", 0
	}

	encoding, params, _ := strings.Cut(value, ";")
	quality := 1.0
	for _, param := range strings.Split(params, ";") {
		name, rawQuality, ok := strings.Cut(strings.TrimSpace(param), "=")
		if !ok || !strings.EqualFold(name, "q") {
			continue
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(rawQuality), 64)
		if err == nil {
			quality = parsed
		} else {
			quality = 0
		}
		break
	}
	if quality < 0 {
		quality = 0
	}
	return strings.ToLower(strings.TrimSpace(encoding)), quality
}

func varyContains(header http.Header, value string) bool {
	for _, existing := range header.Values("Vary") {
		for _, part := range strings.Split(existing, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) {
				return true
			}
		}
	}
	return false
}
