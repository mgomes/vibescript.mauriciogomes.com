package site

import (
	"compress/gzip"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

const gzipCompressionLevel = 5

func gzipResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !acceptsGzip(r.Header.Get("Accept-Encoding")) || r.Method == http.MethodHead || r.Header.Get("Range") != "" {
			next.ServeHTTP(w, r)
			return
		}

		writer := &gzipResponseWriter{ResponseWriter: w}
		defer writer.Close()

		next.ServeHTTP(writer, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	wroteHeader bool
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	if shouldGzip(status, w.Header()) {
		writer, err := gzip.NewWriterLevel(w.ResponseWriter, gzipCompressionLevel)
		if err == nil {
			header := w.Header()
			addVary(header, "Accept-Encoding")
			header.Set("Content-Encoding", "gzip")
			header.Del("Content-Length")
			header.Del("Accept-Ranges")
			w.writer = writer
		}
	}

	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", http.DetectContentType(body))
		}
		w.WriteHeader(http.StatusOK)
	}

	if w.writer != nil {
		return w.writer.Write(body)
	}
	return w.ResponseWriter.Write(body)
}

func (w *gzipResponseWriter) Close() error {
	if w.writer == nil {
		return nil
	}
	return w.writer.Close()
}

func shouldGzip(status int, header http.Header) bool {
	if !statusAllowsBody(status) {
		return false
	}
	if header.Get("Content-Encoding") != "" {
		return false
	}
	return compressibleContentType(header.Get("Content-Type"))
}

func statusAllowsBody(status int) bool {
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

func compressibleContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	}

	switch {
	case strings.HasPrefix(mediaType, "text/"):
		return true
	case mediaType == "application/json":
		return true
	case mediaType == "application/javascript":
		return true
	case mediaType == "application/xml":
		return true
	case mediaType == "image/svg+xml":
		return true
	default:
		return false
	}
}

func acceptsGzip(header string) bool {
	for _, part := range strings.Split(header, ",") {
		token, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		if !strings.EqualFold(token, "gzip") {
			continue
		}
		return encodingQuality(params) > 0
	}
	return false
}

func encodingQuality(params string) float64 {
	for _, param := range strings.Split(params, ";") {
		name, value, ok := strings.Cut(strings.TrimSpace(param), "=")
		if !ok || !strings.EqualFold(name, "q") {
			continue
		}
		quality, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0
		}
		return quality
	}
	return 1
}

func addVary(header http.Header, value string) {
	for _, existing := range header.Values("Vary") {
		for _, part := range strings.Split(existing, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) {
				return
			}
		}
	}
	header.Add("Vary", value)
}
