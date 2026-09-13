package api

import (
	"github.com/freemed/freemed-server/common"
	"github.com/gin-gonic/gin"
)

// maxListLimit is the hard ceiling on one page of a list endpoint, matching the
// clamp api/audit.go applies to GET /api/admin/audit-log. Without it a caller
// could ask for `?limit=1000000` and receive the whole table — the same
// unbounded response M7 is about, only requested explicitly.
const maxListLimit int64 = 200

// maxListOffset bounds the offset before it is narrowed to the int32 the sqlc
// LIMIT/OFFSET parameters use. An unbounded int64 offset would wrap on the
// narrowing conversion and reach MySQL as a negative or nonsense value.
const maxListOffset int64 = 1_000_000

// pageParams reads ?offset= and ?limit= for a bounded list endpoint and returns
// them as the int32 the generated LIMIT/OFFSET parameters take. Callers that
// want the whole page give the query no other bound, so this is the only limit
// on the response size.
//
// It reuses common.ClampPagination for the negative-value repair (a negative
// offset previously reached a slice expression and panicked the handler, and a
// negative LIMIT reached the driver as an error) and then applies the house
// upper bound itself, because ClampPagination deliberately leaves the ceiling
// to the caller.
func pageParams(c *gin.Context) (offset, limit int32) {
	off := common.ParseInt(c.DefaultQuery("offset", "0"))
	lim := common.ParseInt(c.DefaultQuery("limit", "50"))
	off, lim = common.ClampPagination(off, lim)
	if lim > maxListLimit {
		lim = maxListLimit
	}
	if off > maxListOffset {
		off = maxListOffset
	}
	return int32(off), int32(lim)
}
