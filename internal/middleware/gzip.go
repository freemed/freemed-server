// Package middleware — gzip response compression.
//
// This replaces github.com/gin-gonic/contrib/gzip, which is deprecated and
// corrupts every raw-body response. The old middleware wrapped the writer and
// compressed the body, but left the Content-Length header alone — and gin's
// c.Data() sets that header to the UNCOMPRESSED length (render/data.go):
//
//	w.Header().Set("Content-Length", strconv.Itoa(len(r.Data)))
//
// The result was an internally inconsistent response: headers announced, say,
// `Content-Length: 8657` while only 844 compressed bytes were sent. Any client
// that honours Content-Length — i.e. every browser — blocks forever waiting for
// the remaining bytes. curl reports it as exit 18, "transfer closed with
// outstanding read data remaining". That silently broke every endpoint written
// with c.Data(): DICOM instance retrieval, the CCDA/XML exports (staff and
// portal), the CMS-1500 PDF, data_store blobs, signature images and the FHIR
// $document operation.
//
// The fix is to drop Content-Length before the first compressed byte reaches the
// wire, so net/http falls back to chunked transfer encoding.
package middleware

import (
	"compress/gzip"
	"strings"

	"github.com/gin-gonic/gin"
)

// gzipMinLength is the smallest body worth compressing. Below this the framing
// overhead costs more than the saving.
const gzipMinLength = 512

// compressibleTypes are the response content types that benefit from gzip.
// Anything else (already-compressed or raw binary such as application/dicom and
// application/pdf) is passed through untouched: compressing it burns CPU per
// request for little or no gain, and the CDN/proxy in front may already handle it.
var compressibleTypes = []string{
	"application/json",
	"application/xml",
	"application/javascript",
	"application/x-javascript",
	"text/",
	"image/svg+xml",
	"/json",
	"/xml",
	"+json",
	"+xml",
}

// Gzip compresses responses for clients that advertise gzip support.
//
// Unlike the middleware it replaces this one is safe with c.Data(): it removes
// the Content-Length header that gin sets to the uncompressed length, and it
// decides at first write whether the body is worth compressing at all.
func Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !clientAcceptsGzip(c) {
			c.Next()
			return
		}

		gz, err := gzip.NewWriterLevel(c.Writer, gzip.DefaultCompression)
		if err != nil {
			// Compression unavailable: serve the response uncompressed rather
			// than failing the request.
			c.Next()
			return
		}

		w := &gzipResponseWriter{ResponseWriter: c.Writer, gz: gz}
		c.Writer = w

		c.Next()

		// Close flushes the gzip trailer. If the writer never started (a body
		// that turned out not to be worth compressing, or an empty response) this
		// is a no-op and the original writer already emitted the response.
		_ = w.close()
	}
}

func clientAcceptsGzip(c *gin.Context) bool {
	if c.Request == nil {
		return false
	}
	// A handler may legitimately pre-compress its own body.
	if c.Writer.Header().Get("Content-Encoding") != "" {
		return false
	}
	for _, enc := range strings.Split(c.Request.Header.Get("Accept-Encoding"), ",") {
		if strings.EqualFold(strings.TrimSpace(strings.SplitN(enc, ";", 2)[0]), "gzip") {
			return true
		}
	}
	return false
}

// gzipResponseWriter compresses the body once it knows the response is worth
// compressing. Until then it buffers nothing: it forwards writes verbatim.
type gzipResponseWriter struct {
	gin.ResponseWriter
	gz *gzip.Writer

	decided bool // a decision about compressing has been made
	active  bool // compressing this response
}

// WriteHeader defers to gin's writer, which records the status without flushing
// it — headers are only committed on the first real Write, which is what lets the
// compress-or-not decision run before anything reaches the wire.
func (g *gzipResponseWriter) WriteHeader(code int) {
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipResponseWriter) WriteHeaderNow() {
	g.ResponseWriter.WriteHeaderNow()
}

// Write decides once whether to compress, then routes the body accordingly.
func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if !g.decided {
		g.decide(len(b))
	}
	if !g.active {
		return g.ResponseWriter.Write(b)
	}
	// Critical: gin's c.Data() sets Content-Length to the uncompressed length.
	// Once we compress, that header is a lie and the client hangs waiting for
	// bytes that will never arrive. Delete it so net/http switches to chunked.
	g.Header().Del("Content-Length")
	return g.gz.Write(b)
}

func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	return g.Write([]byte(s))
}

func (g *gzipResponseWriter) decide(n int) {
	g.decided = true
	if !compressibleContentType(g.Header().Get("Content-Type")) {
		return
	}
	if g.Header().Get("Content-Length") != "" && n < gzipMinLength {
		// A known small body: leave it alone.
		return
	}
	g.active = true
	g.Header().Set("Content-Encoding", "gzip")
	g.Header().Add("Vary", "Accept-Encoding")
	g.Header().Del("Content-Length")
}

// close finishes the gzip stream if compression started. It reports whether a
// compressed body was written.
func (g *gzipResponseWriter) close() error {
	if !g.active {
		return nil
	}
	return g.gz.Close()
}

// Flush lets streaming handlers keep working when compression is active.
func (g *gzipResponseWriter) Flush() {
	if g.active && g.gz != nil {
		_ = g.gz.Flush()
	}
	g.ResponseWriter.Flush()
}

func compressibleContentType(ct string) bool {
	if ct == "" {
		// Nothing was set: c.Data() always sets one, and gin's JSON/XML renderers
		// do too, so an empty content type means an unusual writer — leave it be
		// rather than guessing.
		return false
	}
	ct = strings.ToLower(ct)
	for _, want := range compressibleTypes {
		if strings.Contains(ct, want) {
			return true
		}
	}
	return false
}
