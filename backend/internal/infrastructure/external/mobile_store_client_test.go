package external

import "testing"

func TestParseAppleReviews_SkipsMetadataEntry(t *testing.T) {
	// Mirrors the real RSS shape: first entry is app metadata (no im:rating), then reviews.
	body := []byte(`{"feed":{"entry":[
		{"im:name":{"label":"App"},"title":{"label":"App Store"}},
		{"author":{"name":{"label":"Rango"}},"im:rating":{"label":"1"},"title":{"label":"Meh"},"content":{"label":"Leave me alone."},"im:version":{"label":"26.30.77"}},
		{"author":{"name":{"label":"Sam"}},"im:rating":{"label":"5"},"title":{"label":"Love it"},"content":{"label":"Great app"},"im:version":{"label":"26.30.77"}}
	]}}`)
	reviews, err := parseAppleReviews(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(reviews) != 2 {
		t.Fatalf("reviews = %d, want 2 (metadata entry skipped)", len(reviews))
	}
	if reviews[0].Author != "Rango" || reviews[0].Rating != 1 || reviews[0].Version != "26.30.77" {
		t.Errorf("review[0] = %+v", reviews[0])
	}
	if reviews[1].Rating != 5 || reviews[1].Title != "Love it" {
		t.Errorf("review[1] = %+v", reviews[1])
	}
}

func TestParseAppleReviews_SingleEntryObject(t *testing.T) {
	// iTunes returns a single object (not an array) when there's one entry.
	body := []byte(`{"feed":{"entry":{"im:name":{"label":"App"},"title":{"label":"App Store"}}}}`)
	reviews, err := parseAppleReviews(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(reviews) != 0 {
		t.Errorf("reviews = %d, want 0 (only the metadata object)", len(reviews))
	}
}

func TestParseGooglePlayLD(t *testing.T) {
	// The real listing embeds one application/ld+json block with aggregateRating.
	html := []byte(`<html><head>
	<script type="application/ld+json" nonce="x">{"@type":"SoftwareApplication","name":"WhatsApp Messenger","image":"https://icon.png","aggregateRating":{"@type":"AggregateRating","ratingValue":"4.633","ratingCount":"240297698"}}</script>
	</head><body>4.2star other app noise</body></html>`)
	s, err := parseGooglePlayLD(html)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if s.AppName != "WhatsApp Messenger" || s.Store != "google_play" {
		t.Errorf("name/store = %q/%q", s.AppName, s.Store)
	}
	if s.RatingValue < 4.6 || s.RatingValue > 4.7 {
		t.Errorf("ratingValue = %v, want ~4.633", s.RatingValue)
	}
	if s.RatingCount != 240297698 {
		t.Errorf("ratingCount = %d, want 240297698", s.RatingCount)
	}
}

func TestParseGooglePlayInstalls(t *testing.T) {
	// Mirrors the real listing: the coarse range precedes the single "Downloads" stat label.
	html := []byte(`<div aria-label="More info">info</div><div><span>10B+</span></div><div class="g1rdde">Downloads</div>`)
	if got := parseGooglePlayInstalls(html); got != "10B+" {
		t.Errorf("installs = %q, want 10B+", got)
	}
	// Long form + a nearer package example.
	html2 := []byte(`<span>500K+</span></div><div>Downloads</div>`)
	if got := parseGooglePlayInstalls(html2); got != "500K+" {
		t.Errorf("installs = %q, want 500K+", got)
	}
	// Best-effort: no "Downloads" label → empty (never errors).
	if got := parseGooglePlayInstalls([]byte(`<html>no stat</html>`)); got != "" {
		t.Errorf("installs = %q, want empty", got)
	}
}

func TestParseGooglePlayLD_NoRating(t *testing.T) {
	if _, err := parseGooglePlayLD([]byte(`<html>no ld json</html>`)); err == nil {
		t.Error("expected an error when the listing has no rating block")
	}
}
