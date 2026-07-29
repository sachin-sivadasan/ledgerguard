package handler

import (
	"net/http"
	"strconv"
)

// maxReportPageLimit bounds a single report-table page so a whale store can never
// force an unbounded payload, even if a client asks for a huge limit.
const maxReportPageLimit = 200

// parsePaging reads ?limit= and ?offset= for a report's list table.
//   - limit <= 0  → "no paging" (the caller returns every row, e.g. CSV export or
//     an older client that doesn't page).
//   - limit > max → clamped to maxReportPageLimit.
//   - offset < 0  → clamped to 0.
//
// The report page requests a small preview (e.g. limit=8); the dedicated detail
// page requests a full window (e.g. limit=50&offset=N). KPIs/trends are always
// computed over the full data set, independent of paging.
func parsePaging(r *http.Request) (limit, offset int) {
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	if limit < 0 {
		limit = 0
	}
	if limit > maxReportPageLimit {
		limit = maxReportPageLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// pageSlice returns the [offset, offset+limit) window of s, bounds-safe. When
// limit <= 0 it returns s unchanged (no paging). An offset past the end yields an
// empty (non-nil) slice so JSON still serializes [] rather than null.
func pageSlice[T any](s []T, offset, limit int) []T {
	if limit <= 0 {
		return s
	}
	if offset >= len(s) {
		return s[len(s):]
	}
	end := min(offset+limit, len(s))
	return s[offset:end]
}
