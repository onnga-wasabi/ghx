package api

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
	"github.com/onnga-wasabi/ghx/internal/model"
)

const issueSearchQuery = `
query($query: String!, $first: Int!) {
  search(query: $query, type: ISSUE, first: $first) {
    nodes {
      ... on Issue {
        number
        title
        body
        state
        url
        createdAt
        updatedAt
        author { login }
        labels(first: 10) { nodes { name } }
        assignees(first: 5) { nodes { login } }
        comments { totalCount }
        repository { nameWithOwner }
      }
    }
  }
}
`

func (c *Client) CloseIssue(ctx context.Context, owner, repo string, number int) error {
	_, _, err := c.GH.Issues.Edit(ctx, owner, repo, number, &github.IssueRequest{
		State: github.Ptr("closed"),
	})
	return err
}

func (c *Client) ReopenIssue(ctx context.Context, owner, repo string, number int) error {
	_, _, err := c.GH.Issues.Edit(ctx, owner, repo, number, &github.IssueRequest{
		State: github.Ptr("open"),
	})
	return err
}

func (c *Client) SearchIssues(ctx context.Context, query string, limit int) ([]model.Issue, error) {
	var data struct {
		Search struct {
			Nodes []issueNode `json:"nodes"`
		} `json:"search"`
	}

	err := c.graphQL(ctx, issueSearchQuery, map[string]interface{}{
		"query": query,
		"first": limit,
	}, &data)
	if err != nil {
		return nil, fmt.Errorf("search issues: %w", err)
	}

	issues := make([]model.Issue, 0, len(data.Search.Nodes))
	for _, n := range data.Search.Nodes {
		if n.Number > 0 {
			issues = append(issues, n.toModel())
		}
	}
	return issues, nil
}

type issueNode struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	URL       string `json:"url"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Labels struct {
		Nodes []struct{ Name string } `json:"nodes"`
	} `json:"labels"`
	Assignees struct {
		Nodes []struct{ Login string } `json:"nodes"`
	} `json:"assignees"`
	Comments struct {
		TotalCount int `json:"totalCount"`
	} `json:"comments"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
}

func (n issueNode) toModel() model.Issue {
	issue := model.Issue{
		Number:   n.Number,
		Title:    n.Title,
		Body:     n.Body,
		State:    n.State,
		URL:      n.URL,
		Author:   n.Author.Login,
		Comments: n.Comments.TotalCount,
		RepoName: n.Repository.NameWithOwner,
	}
	for _, l := range n.Labels.Nodes {
		issue.Labels = append(issue.Labels, l.Name)
	}
	for _, a := range n.Assignees.Nodes {
		issue.Assignees = append(issue.Assignees, a.Login)
	}
	return issue
}
