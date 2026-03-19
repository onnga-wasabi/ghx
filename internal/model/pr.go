package model

import "time"

type PR struct {
	Number      int
	Title       string
	Body        string
	State       string
	Author      string
	HeadRef     string
	BaseRef     string
	URL         string
	IsDraft     bool
	Mergeable   string
	Additions   int
	Deletions   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	RepoName    string
	Checks      []Check
	Labels      []string
	ReviewState string
}

type Check struct {
	Name       string
	Status     string
	Conclusion string
	URL        string
}

func (p PR) StatusIcon() string {
	if p.IsDraft {
		return "◇"
	}
	switch p.State {
	case "OPEN":
		return "●"
	case "MERGED":
		return "◆"
	case "CLOSED":
		return "○"
	default:
		return "?"
	}
}

func (p PR) ChecksSummary() (pass, fail, pending int) {
	for _, c := range p.Checks {
		switch c.Conclusion {
		case "SUCCESS":
			pass++
		case "FAILURE", "TIMED_OUT", "ACTION_REQUIRED":
			fail++
		default:
			if c.Status != "COMPLETED" {
				pending++
			}
		}
	}
	return
}
