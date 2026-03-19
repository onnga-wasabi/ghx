package model

import "time"

type Job struct {
	ID          int64
	RunID       int64
	Name        string
	Status      string
	Conclusion  string
	StartedAt   time.Time
	CompletedAt time.Time
	Steps       []Step
	HTMLURL     string
}

type Step struct {
	Name       string
	Status     string
	Conclusion string
	Number     int64
}

func (j Job) StatusIcon() string {
	switch j.Status {
	case "completed":
		switch j.Conclusion {
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
	case "in_progress":
		return "⏳"
	default:
		return "🕐"
	}
}
