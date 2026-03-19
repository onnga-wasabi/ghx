package model

import "time"

type Issue struct {
	Number    int
	Title     string
	Body      string
	State     string
	Author    string
	URL       string
	Labels    []string
	Assignees []string
	CreatedAt time.Time
	UpdatedAt time.Time
	RepoName  string
	Comments  int
}

func (i Issue) StatusIcon() string {
	switch i.State {
	case "OPEN":
		return "●"
	case "CLOSED":
		return "✓"
	default:
		return "?"
	}
}
