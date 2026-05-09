package entity

// CohortData represents retention data for a single cohort (month of subscription creation).
type CohortData struct {
	CohortMonth   string    `json:"cohort_month"`   // e.g. "2025-10"
	InitialStores int       `json:"initial_stores"`
	RetentionPcts []float64 `json:"retention_pcts"` // M0=100%, M1, M2, ...
}
