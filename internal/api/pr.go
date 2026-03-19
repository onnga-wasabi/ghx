package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-github/v68/github"
	"github.com/onnga-wasabi/ghx/internal/model"
)

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *Client) graphQL(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	body, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/graphql", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return err
	}
	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("graphql: %s", gqlResp.Errors[0].Message)
	}
	return json.Unmarshal(gqlResp.Data, result)
}

const prSearchQuery = `
query($query: String!, $first: Int!) {
  search(query: $query, type: ISSUE, first: $first) {
    nodes {
      ... on PullRequest {
        number
        title
        body
        state
        isDraft
        mergeable
        additions
        deletions
        headRefName
        baseRefName
        url
        createdAt
        updatedAt
        author { login }
        labels(first: 10) { nodes { name } }
        reviewDecision
        statusCheckRollup {
          contexts(first: 50) {
            nodes {
              ... on CheckRun {
                name
                status
                conclusion
                detailsUrl
              }
              ... on StatusContext {
                context
                state
                targetUrl
              }
            }
          }
        }
        repository { nameWithOwner }
      }
    }
  }
}
`

func (c *Client) ApprovePR(ctx context.Context, owner, repo string, number int) error {
	_, _, err := c.GH.PullRequests.CreateReview(ctx, owner, repo, number, &github.PullRequestReviewRequest{
		Event: github.Ptr("APPROVE"),
	})
	return err
}

func (c *Client) MergePR(ctx context.Context, owner, repo string, number int) error {
	_, _, err := c.GH.PullRequests.Merge(ctx, owner, repo, number, "", nil)
	return err
}

func (c *Client) ClosePR(ctx context.Context, owner, repo string, number int) error {
	_, _, err := c.GH.PullRequests.Edit(ctx, owner, repo, number, &github.PullRequest{
		State: github.Ptr("closed"),
	})
	return err
}

func (c *Client) SearchPRs(ctx context.Context, query string, limit int) ([]model.PR, error) {
	var data struct {
		Search struct {
			Nodes []prNode `json:"nodes"`
		} `json:"search"`
	}

	err := c.graphQL(ctx, prSearchQuery, map[string]interface{}{
		"query": query,
		"first": limit,
	}, &data)
	if err != nil {
		return nil, fmt.Errorf("search PRs: %w", err)
	}

	prs := make([]model.PR, 0, len(data.Search.Nodes))
	for _, n := range data.Search.Nodes {
		prs = append(prs, n.toModel())
	}
	return prs, nil
}

type prNode struct {
	Number      int      `json:"number"`
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	State       string   `json:"state"`
	IsDraft     bool     `json:"isDraft"`
	Mergeable   string   `json:"mergeable"`
	Additions   int      `json:"additions"`
	Deletions   int      `json:"deletions"`
	HeadRefName string   `json:"headRefName"`
	BaseRefName string   `json:"baseRefName"`
	URL         string   `json:"url"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	Author      struct {
		Login string `json:"login"`
	} `json:"author"`
	Labels struct {
		Nodes []struct{ Name string } `json:"nodes"`
	} `json:"labels"`
	ReviewDecision    string `json:"reviewDecision"`
	StatusCheckRollup *struct {
		Contexts struct {
			Nodes []checkNode `json:"nodes"`
		} `json:"contexts"`
	} `json:"statusCheckRollup"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
}

type checkNode struct {
	Name       string `json:"name"`
	Context    string `json:"context"`
	Status     string `json:"status"`
	State      string `json:"state"`
	Conclusion string `json:"conclusion"`
	DetailsURL string `json:"detailsUrl"`
	TargetURL  string `json:"targetUrl"`
}

func (n prNode) toModel() model.PR {
	pr := model.PR{
		Number:      n.Number,
		Title:       n.Title,
		Body:        n.Body,
		State:       n.State,
		IsDraft:     n.IsDraft,
		Mergeable:   n.Mergeable,
		Additions:   n.Additions,
		Deletions:   n.Deletions,
		HeadRef:     n.HeadRefName,
		BaseRef:     n.BaseRefName,
		URL:         n.URL,
		Author:      n.Author.Login,
		ReviewState: n.ReviewDecision,
		RepoName:    n.Repository.NameWithOwner,
	}

	for _, l := range n.Labels.Nodes {
		pr.Labels = append(pr.Labels, l.Name)
	}

	if n.StatusCheckRollup != nil {
		for _, c := range n.StatusCheckRollup.Contexts.Nodes {
			check := model.Check{URL: c.DetailsURL}
			if c.Name != "" {
				check.Name = c.Name
				check.Status = c.Status
				check.Conclusion = c.Conclusion
			} else {
				check.Name = c.Context
				check.Status = c.State
				check.URL = c.TargetURL
			}
			pr.Checks = append(pr.Checks, check)
		}
	}

	return pr
}
