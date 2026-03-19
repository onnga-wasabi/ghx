package api

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/v68/github"
	"github.com/onnga-wasabi/ghx/internal/model"
)

func (c *Client) ListWorkflows(ctx context.Context, owner, repo string) ([]model.Workflow, error) {
	var all []model.Workflow
	opts := &github.ListOptions{PerPage: 100}
	for {
		result, resp, err := c.GH.Actions.ListWorkflows(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("list workflows: %w", err)
		}
		for _, w := range result.Workflows {
			all = append(all, model.Workflow{
				ID:    w.GetID(),
				Name:  w.GetName(),
				Path:  w.GetPath(),
				State: w.GetState(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}

func (c *Client) ListRuns(ctx context.Context, owner, repo string, workflowID int64) ([]model.Run, error) {
	return c.ListRunsWithBranch(ctx, owner, repo, workflowID, "")
}

func (c *Client) ListRunsWithBranch(ctx context.Context, owner, repo string, workflowID int64, branch string) ([]model.Run, error) {
	opts := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 50},
	}
	if branch != "" {
		opts.Branch = branch
	}

	var result *github.WorkflowRuns
	var err error

	if workflowID > 0 {
		result, _, err = c.GH.Actions.ListWorkflowRunsByID(ctx, owner, repo, workflowID, opts)
	} else {
		result, _, err = c.GH.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, opts)
	}
	if err != nil {
		return nil, fmt.Errorf("list runs: %w", err)
	}

	runs := make([]model.Run, 0, len(result.WorkflowRuns))
	for _, r := range result.WorkflowRuns {
		runs = append(runs, convertRun(r))
	}
	return runs, nil
}

func (c *Client) ListJobs(ctx context.Context, owner, repo string, runID int64) ([]model.Job, error) {
	opts := &github.ListWorkflowJobsOptions{
		Filter:      "latest",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	result, _, err := c.GH.Actions.ListWorkflowJobs(ctx, owner, repo, runID, opts)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}

	jobs := make([]model.Job, 0, len(result.Jobs))
	for _, j := range result.Jobs {
		jobs = append(jobs, convertJob(j))
	}
	return jobs, nil
}

func (c *Client) GetJobLogs(ctx context.Context, owner, repo string, jobID int64) (string, error) {
	url, _, err := c.GH.Actions.GetWorkflowJobLogs(ctx, owner, repo, jobID, 2)
	if err != nil {
		return "", fmt.Errorf("get job logs URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create log request: %w", err)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch logs: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read logs: %w", err)
	}
	return string(body), nil
}

func (c *Client) RerunWorkflow(ctx context.Context, owner, repo string, runID int64) error {
	_, err := c.GH.Actions.RerunWorkflowByID(ctx, owner, repo, runID)
	if err != nil {
		return fmt.Errorf("rerun workflow: %w", err)
	}
	return nil
}

func (c *Client) RerunFailedJobs(ctx context.Context, owner, repo string, runID int64) error {
	_, err := c.GH.Actions.RerunFailedJobsByID(ctx, owner, repo, runID)
	if err != nil {
		return fmt.Errorf("rerun failed jobs: %w", err)
	}
	return nil
}

func (c *Client) CancelRun(ctx context.Context, owner, repo string, runID int64) error {
	_, err := c.GH.Actions.CancelWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return fmt.Errorf("cancel run: %w", err)
	}
	return nil
}

func (c *Client) TriggerWorkflow(ctx context.Context, owner, repo, workflowFile, ref string, inputs map[string]interface{}) error {
	event := github.CreateWorkflowDispatchEventRequest{
		Ref:    ref,
		Inputs: inputs,
	}
	_, err := c.GH.Actions.CreateWorkflowDispatchEventByFileName(ctx, owner, repo, workflowFile, event)
	if err != nil {
		return fmt.Errorf("trigger workflow: %w", err)
	}
	return nil
}

func convertRun(r *github.WorkflowRun) model.Run {
	run := model.Run{
		ID:         r.GetID(),
		Name:       r.GetName(),
		Status:     r.GetStatus(),
		Conclusion: r.GetConclusion(),
		HeadBranch: r.GetHeadBranch(),
		HeadSHA:    r.GetHeadSHA(),
		Event:      r.GetEvent(),
		WorkflowID: r.GetWorkflowID(),
		RunNumber:  r.GetRunNumber(),
		RunAttempt: r.GetRunAttempt(),
		HTMLURL:    r.GetHTMLURL(),
	}
	if r.CreatedAt != nil {
		run.CreatedAt = r.CreatedAt.Time
	}
	if r.UpdatedAt != nil {
		run.UpdatedAt = r.UpdatedAt.Time
	}
	return run
}

func convertJob(j *github.WorkflowJob) model.Job {
	job := model.Job{
		ID:         j.GetID(),
		RunID:      j.GetRunID(),
		Name:       j.GetName(),
		Status:     j.GetStatus(),
		Conclusion: j.GetConclusion(),
		HTMLURL:    j.GetHTMLURL(),
	}
	if j.StartedAt != nil {
		job.StartedAt = j.StartedAt.Time
	}
	if j.CompletedAt != nil {
		job.CompletedAt = j.CompletedAt.Time
	}
	for _, s := range j.Steps {
		job.Steps = append(job.Steps, model.Step{
			Name:       s.GetName(),
			Status:     s.GetStatus(),
			Conclusion: s.GetConclusion(),
			Number:     s.GetNumber(),
		})
	}
	return job
}
