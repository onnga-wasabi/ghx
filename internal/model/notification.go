package model

import "time"

type Notification struct {
	ID         string
	Title      string
	Type       string // PullRequest, Issue, Release, etc.
	Reason     string // author, mention, review_requested, etc.
	Unread     bool
	RepoName   string
	URL        string
	HTMLURL    string
	UpdatedAt  time.Time
}

func (n Notification) TypeIcon() string {
	switch n.Type {
	case "PullRequest":
		return "⑂"
	case "Issue":
		return "●"
	case "Release":
		return "◆"
	case "Discussion":
		return "💬"
	default:
		return "•"
	}
}
