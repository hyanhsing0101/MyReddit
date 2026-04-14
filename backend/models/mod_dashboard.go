package models

type ModerationDashboardData struct {
	PendingReports    int64 `json:"pending_reports"`
	InReviewReports   int64 `json:"in_review_reports"`
	ResolvedReports7d int64 `json:"resolved_reports_7d"`
	RejectedReports7d int64 `json:"rejected_reports_7d"`
	Logs24h           int64 `json:"logs_24h"`
	ReportsCreated24h int64 `json:"reports_created_24h"`
}
