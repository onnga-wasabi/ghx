package model

import "time"

type Run struct {
	ID           int64
	Name         string
	Status       string // queued, in_progress, completed, waiting, requested, pending
	Conclusion   string // success, failure, cancelled, skipped, timed_out, action_required
	HeadBranch   string
	HeadSHA      string
	Event        string
	WorkflowID   int64
	RunNumber    int
	RunAttempt   int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	HTMLURL      string
}

func (r Run) IsCompleted() bool {
	return r.Status == "completed"
}

func (r Run) IsFailure() bool {
	return r.Conclusion == "failure" || r.Conclusion == "timed_out"
}

func (r Run) StatusIcon() string {
	if !r.IsCompleted() {
		switch r.Status {
		case "in_progress":
			return "⏳"
		case "queued", "waiting", "pending", "requested":
			return "🕐"
		}
		return "⏳"
	}
	switch r.Conclusion {
	case "success":
		return "✓"
	case "failure":
		return "✗"
	case "cancelled":
		return "⊘"
	case "skipped":
		return "⊘"
	default:
		return "?"
	}
}
