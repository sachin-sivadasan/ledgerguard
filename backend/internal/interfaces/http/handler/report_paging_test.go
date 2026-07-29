package handler

import (
	"net/http/httptest"
	"testing"
)

func TestParsePaging(t *testing.T) {
	cases := []struct {
		name       string
		query      string
		wantLimit  int
		wantOffset int
	}{
		{"empty means no paging", "", 0, 0},
		{"in-range", "limit=50&offset=100", 50, 100},
		{"limit capped at max", "limit=99999", maxReportPageLimit, 0},
		{"limit exactly at cap", "limit=200", 200, 0},
		{"negative limit clamps to 0", "limit=-5", 0, 0},
		{"negative offset clamps to 0", "offset=-10&limit=20", 20, 0},
		{"non-numeric is treated as unpaged/zero", "limit=abc&offset=xyz", 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/x?"+c.query, nil)
			limit, offset := parsePaging(req)
			if limit != c.wantLimit || offset != c.wantOffset {
				t.Errorf("parsePaging(%q) = (limit %d, offset %d), want (%d, %d)",
					c.query, limit, offset, c.wantLimit, c.wantOffset)
			}
		})
	}
}

func TestPageSlice(t *testing.T) {
	s := []int{0, 1, 2, 3, 4}

	// limit <= 0 → whole slice unchanged (CSV / older-client path).
	if got := pageSlice(s, 2, 0); len(got) != 5 {
		t.Errorf("limit=0: got %v, want all 5", got)
	}

	// mid-window.
	if got := pageSlice(s, 1, 2); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Errorf("offset=1,limit=2: got %v, want [1 2]", got)
	}

	// limit overshoots the tail → clamped to end.
	if got := pageSlice(s, 3, 10); len(got) != 2 || got[0] != 3 || got[1] != 4 {
		t.Errorf("offset=3,limit=10: got %v, want [3 4]", got)
	}

	// offset past the end → empty but NON-NIL (so JSON serializes [] not null).
	got := pageSlice(s, 9, 5)
	if len(got) != 0 {
		t.Errorf("offset past end: got %v, want empty", got)
	}
	if got == nil {
		t.Error("offset past end returned nil; must be non-nil for [] JSON")
	}
}
