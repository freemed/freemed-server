package middleware

import (
	"net/http"
	"strconv"

	"github.com/freemed/freemed-server/common"
	"github.com/gin-gonic/gin"
)

// MaxRequestBodyBytes is the default cap applied to request bodies on the /api
// group: 32 MiB.
//
// Assumed maximum legitimate body: the largest payloads this server accepts are
// raw DICOM objects (POST /api/patient/:id/dicom, both the multipart and the
// base64 JSON fallback) and the text ingest endpoints that read the whole body
// with io.ReadAll — HL7 (api/hl7.go), ERA 835 (api/era.go), CCDA (api/ccda.go)
// and eligibility 270/271 (api/eligibility.go). A single DICOM instance with
// encapsulated documents is normally well under 10 MiB and an HL7 message is
// kilobytes, so 32 MiB leaves generous headroom while staying far below the
// 315 MB body that previously drove the process RSS from 30 MB to 1.71 GB.
// Note this also bounds the depth-recursion attack surface of the DICOM parser
// (see pkg/dicom/parser.go MaxSequenceDepth).
const MaxRequestBodyBytes = 32 << 20

// MaxBody returns a Gin middleware that rejects request bodies larger than
// limit bytes. Requests whose declared Content-Length exceeds the limit are
// refused up front with 413 and a standard JSON error body; bodies of unknown
// length (chunked transfer encoding) are still capped in memory by
// http.MaxBytesReader, so the limit holds even when a client lies about or omits
// the length. In that chunked case the handler's read fails with
// *http.MaxBytesError and it answers with its own error (400) rather than 413 —
// the payload is still refused, only the status code is less precise.
func MaxBody(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limit <= 0 {
			c.Next()
			return
		}

		if c.Request != nil && c.Request.ContentLength > limit {
			common.ErrorResponse(c, http.StatusRequestEntityTooLarge,
				"request body too large: limit is "+strconv.FormatInt(limit, 10)+" bytes")
			return
		}

		if c.Request != nil && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}

		c.Next()
	}
}
