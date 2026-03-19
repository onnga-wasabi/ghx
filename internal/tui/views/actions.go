package views

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/onnga-wasabi/ghx/internal/api"
	"github.com/onnga-wasabi/ghx/internal/model"
	"github.com/onnga-wasabi/ghx/internal/tui/components"
	"github.com/onnga-wasabi/ghx/internal/tui/keys"
	"github.com/onnga-wasabi/ghx/internal/tui/styles"
)

type actionsMsg struct {
	workflows []model.Workflow
	err       error
}
type runsMsg struct {
	runs []model.Run
	err  error
}
type jobsMsg struct {
	jobs []model.Job
	err  error
}
type logsMsg struct {
	logs string
	err  error
}
type actionResultMsg struct {
	action string
	err    error
}
type triggerInputMsg struct {
	ref    string
	inputs map[string]interface{}
}
type toggleJobsMsg struct {
	runID int64
	jobs  []model.Job
	err   error
}

type runTableEntry struct {
	isRun bool
	runID int64
	jobID int64
}

type statusFilter struct {
	label  string
	filter string // "" = all, "success", "failure", "in_progress"
}

var statusFilters = []statusFilter{
	{label: "All", filter: ""},
	{label: "✓ Success", filter: "success"},
	{label: "✗ Failed", filter: "failure"},
	{label: "⏳ Running", filter: "in_progress"},
}

const actionsFilterBarH = 1

type ActionsView struct {
	client *api.Client
	owner  string
	repo   string

	workflows []model.Workflow
	allRuns   []model.Run
	runs      []model.Run
	jobs      []model.Job
	wfTable   *components.Table
	runTable  *components.Table
	jobTable  *components.Table
	logViewer *components.LogViewer

	expandedRuns    map[int64]bool
	runJobs         map[int64][]model.Job
	runTableEntries []runTableEntry

	statusIdx     int
	selectedWfIdx int
	filterQuery   string

	focusPane   int // 0=workflows, 1=runs, 2=jobs, 3=logs
	fullscreen  bool
	showLogs    bool
	smartLayout bool

	loading    bool
	loadingMsg string
	spinner    spinner.Model
	err        error

	showConfirm bool
	confirmMsg  string
	confirmFn   func() tea.Cmd

	triggerMode     bool
	triggerRefInput string

	width  int
	height int
}

func NewActionsView(client *api.Client, owner, repo string, smartLayout bool) *ActionsView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return &ActionsView{
		client:       client,
		owner:        owner,
		repo:         repo,
		wfTable:      components.NewTable("Workflows"),
		runTable:     components.NewTable("Runs"),
		jobTable:     components.NewTable("Jobs"),
		logViewer:    components.NewLogViewer(),
		spinner:      s,
		showLogs:     true,
		smartLayout:  smartLayout,
		expandedRuns: make(map[int64]bool),
		runJobs:      make(map[int64][]model.Job),
	}
}

func (v *ActionsView) Name() string { return "Actions" }

func (v *ActionsView) WantsFilterKey() bool {
	return v.focusPane == 3 && v.showLogs
}

func (v *ActionsView) IsInputMode() bool {
	return v.logViewer.IsSearching()
}

func (v *ActionsView) ShowRunsForBranch(branch string) tea.Cmd {
	v.loading = true
	v.loadingMsg = fmt.Sprintf("Loading runs for %s...", branch)
	v.runs = nil
	v.runTable.ClearItems("Loading…")
	v.runTable.Title = fmt.Sprintf("Runs (%s)", branch)
	v.jobs = nil
	v.jobTable.ClearItems("")
	v.logViewer.SetContent("")
	v.focusPane = 1
	v.expandedRuns = make(map[int64]bool)
	v.runTableEntries = nil
	v.recalcLayout()
	return func() tea.Msg {
		runs, err := v.client.ListRunsWithBranch(context.Background(), v.owner, v.repo, 0, branch)
		return runsMsg{runs: runs, err: err}
	}
}

func (v *ActionsView) KeyMap() help.KeyMap { return keys.Actions }

func (v *ActionsView) Init() tea.Cmd {
	return tea.Batch(v.fetchWorkflows(), v.spinner.Tick)
}

func (v *ActionsView) SetSize(w, h int) {
	v.width = w
	v.height = h
	v.recalcLayout()
}

func (v *ActionsView) paneHeight() int {
	return max(4, v.height-actionsFilterBarH)
}

func (v *ActionsView) recalcLayout() {
	if v.fullscreen {
		v.logViewer.SetSize(v.width, v.height)
		v.logViewer.Active = true
		return
	}

	ph := v.paneHeight()
	var wfW, runW, jobW, logW int

	if !v.showLogs {
		if v.smartLayout {
			const shrunk = 10
			switch v.focusPane {
			case 0:
				wfW = v.width * 30 / 100
				runW = v.width * 35 / 100
				jobW = v.width - wfW - runW
			case 1:
				wfW = v.width * shrunk / 100
				runW = v.width * 50 / 100
				jobW = v.width - wfW - runW
			case 2:
				wfW = v.width * shrunk / 100
				runW = v.width * 20 / 100
				jobW = v.width - wfW - runW
			default:
				wfW = v.width * 25 / 100
				runW = v.width * 35 / 100
				jobW = v.width - wfW - runW
			}
		} else {
			wfW = v.width * 25 / 100
			runW = v.width * 35 / 100
			jobW = v.width - wfW - runW
		}
	} else if v.smartLayout {
		const shrunk = 10
		switch v.focusPane {
		case 0:
			wfW = v.width * 25 / 100
			runW = v.width * 25 / 100
			jobW = v.width * 20 / 100
			logW = v.width - wfW - runW - jobW
		case 1:
			wfW = v.width * shrunk / 100
			runW = v.width * 35 / 100
			jobW = v.width * 20 / 100
			logW = v.width - wfW - runW - jobW
		case 2:
			wfW = v.width * shrunk / 100
			runW = v.width * 15 / 100
			jobW = v.width * 30 / 100
			logW = v.width - wfW - runW - jobW
		case 3:
			wfW = v.width * shrunk / 100
			runW = v.width * shrunk / 100
			jobW = v.width * 12 / 100
			logW = v.width - wfW - runW - jobW
		}
	} else {
		wfW = v.width * 20 / 100
		runW = v.width * 25 / 100
		jobW = v.width * 20 / 100
		logW = v.width - wfW - runW - jobW
	}

	v.wfTable.Width = wfW
	v.wfTable.Height = ph
	v.runTable.Width = runW
	v.runTable.Height = ph
	v.jobTable.Width = jobW
	v.jobTable.Height = ph
	if v.showLogs {
		v.logViewer.SetSize(logW, ph)
	}

	v.wfTable.Active = v.focusPane == 0
	v.runTable.Active = v.focusPane == 1
	v.jobTable.Active = v.focusPane == 2
	v.logViewer.Active = v.focusPane == 3 && v.showLogs
}

func (v *ActionsView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case actionsMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		allEntry := model.Workflow{ID: 0, Name: "All Workflows", State: "active"}
		v.workflows = append([]model.Workflow{allEntry}, msg.workflows...)
		v.updateWorkflowTable()
		return v, v.fetchRuns(0)

	case runsMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.allRuns = msg.runs
		v.applyStatusFilter()
		if len(v.runs) > 0 {
			return v, v.fetchJobs(v.runs[0].ID)
		}

	case jobsMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.jobs = msg.jobs
		v.updateJobTable()
		if len(v.jobs) > 0 {
			return v, v.fetchLogs(v.jobs[0].ID)
		}

	case logsMsg:
		v.loading = false
		if msg.err != nil {
			v.logViewer.SetContent("Error loading logs: " + msg.err.Error())
			return v, nil
		}
		v.logViewer.SetContent(msg.logs)

	case actionResultMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		return v, v.refreshCurrent()

	case toggleJobsMsg:
		if msg.err == nil {
			v.runJobs[msg.runID] = msg.jobs
			v.updateRunTable()
		}

	case tea.KeyMsg:
		if v.showConfirm {
			return v.handleConfirm(msg)
		}
		return v.handleKey(msg)
	}

	if v.focusPane == 3 {
		var cmd tea.Cmd
		v.logViewer, cmd = v.logViewer.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v *ActionsView) handleKey(msg tea.KeyMsg) (View, tea.Cmd) {
	if v.logViewer.IsSearching() {
		var cmd tea.Cmd
		v.logViewer, cmd = v.logViewer.Update(msg)
		return v, cmd
	}

	switch {
	case key.Matches(msg, keys.Global.Left):
		if v.fullscreen {
			v.fullscreen = false
			v.recalcLayout()
			return v, nil
		}
		if v.focusPane > 0 {
			v.focusPane--
			v.recalcLayout()
		}
		return v, nil
	case key.Matches(msg, keys.Global.Right):
		maxPane := 3
		if !v.showLogs {
			maxPane = 2
		}
		if v.focusPane < maxPane {
			v.focusPane++
			v.recalcLayout()
		}
		return v, nil
	case key.Matches(msg, keys.Global.Enter):
		switch v.focusPane {
		case 1:
			return v, v.toggleRunExpansion()
		case 3:
			v.logViewer.ToggleCurrentSection()
			return v, nil
		}
		return v, nil
	case key.Matches(msg, keys.Actions.StatusToggle):
		v.statusIdx = (v.statusIdx + 1) % len(statusFilters)
		v.applyStatusFilter()
		return v, nil
	case key.Matches(msg, keys.Global.Refresh):
		return v, v.fetchWorkflows()
	case key.Matches(msg, keys.Global.Open):
		return v, v.openInBrowser()
	case key.Matches(msg, keys.Actions.Rerun):
		return v, v.rerunSelected()
	case key.Matches(msg, keys.Actions.RerunFailed):
		return v, v.rerunFailedSelected()
	case key.Matches(msg, keys.Actions.Cancel):
		v.requestConfirm("Cancel this run?", v.cancelSelected)
		return v, nil
	case key.Matches(msg, keys.Actions.Trigger):
		return v, v.triggerSelected()
	case key.Matches(msg, keys.Actions.Fullscreen):
		v.fullscreen = !v.fullscreen
		if v.fullscreen {
			v.showLogs = true
		}
		v.recalcLayout()
		return v, nil
	case key.Matches(msg, keys.Actions.LogToggle):
		v.showLogs = !v.showLogs
		if !v.showLogs && v.focusPane == 3 {
			v.focusPane = 2
		}
		v.recalcLayout()
		return v, nil
	case key.Matches(msg, keys.Actions.WordWrap):
		v.logViewer.ToggleWordWrap()
		return v, nil
	case key.Matches(msg, keys.Actions.SearchNext):
		v.logViewer.NextMatch()
		return v, nil
	case key.Matches(msg, keys.Actions.SearchPrev):
		v.logViewer.PrevMatch()
		return v, nil
	}

	if v.focusPane == 3 {
		if v.logViewer.IsSearching() {
			var cmd tea.Cmd
			v.logViewer, cmd = v.logViewer.Update(msg)
			return v, cmd
		}
		switch {
		case key.Matches(msg, keys.Global.Filter):
			return v, v.logViewer.StartSearch()
		case key.Matches(msg, keys.Actions.ExpandAll):
			v.logViewer.ExpandAll()
			return v, nil
		case key.Matches(msg, keys.Actions.CollapseAll):
			v.logViewer.CollapseAll()
			return v, nil
		case key.Matches(msg, keys.Actions.PrevSection):
			v.logViewer.PrevSection()
			return v, nil
		case key.Matches(msg, keys.Actions.NextSection):
			v.logViewer.NextSection()
			return v, nil
		}
		var cmd tea.Cmd
		v.logViewer, cmd = v.logViewer.Update(msg)
		return v, cmd
	}

	switch {
	case key.Matches(msg, keys.Global.Up):
		v.activeTable().MoveUp()
		return v, v.onSelectionChange()
	case key.Matches(msg, keys.Global.Down):
		v.activeTable().MoveDown()
		return v, v.onSelectionChange()
	case key.Matches(msg, keys.Global.FirstLine):
		v.activeTable().GoToFirst()
		return v, v.onSelectionChange()
	case key.Matches(msg, keys.Global.LastLine):
		v.activeTable().GoToLast()
		return v, v.onSelectionChange()
	}
	return v, nil
}

func (v *ActionsView) handleConfirm(msg tea.KeyMsg) (View, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		v.showConfirm = false
		if v.confirmFn != nil {
			return v, v.confirmFn()
		}
	case "n", "N", "esc", "q":
		v.showConfirm = false
	}
	return v, nil
}

func (v *ActionsView) requestConfirm(msg string, fn func() tea.Cmd) {
	v.showConfirm = true
	v.confirmMsg = msg
	v.confirmFn = fn
}

func (v *ActionsView) activeTable() *components.Table {
	switch v.focusPane {
	case 1:
		return v.runTable
	case 2:
		return v.jobTable
	default:
		return v.wfTable
	}
}

func (v *ActionsView) selectedRunTableEntry() *runTableEntry {
	if v.runTable.Cursor >= 0 && v.runTable.Cursor < len(v.runTableEntries) {
		return &v.runTableEntries[v.runTable.Cursor]
	}
	return nil
}

func (v *ActionsView) toggleRunExpansion() tea.Cmd {
	entry := v.selectedRunTableEntry()
	if entry == nil || !entry.isRun {
		return nil
	}
	runID := entry.runID
	if v.expandedRuns[runID] {
		delete(v.expandedRuns, runID)
		v.updateRunTable()
		return nil
	}
	v.expandedRuns[runID] = true
	if _, ok := v.runJobs[runID]; ok {
		v.updateRunTable()
		return nil
	}
	return func() tea.Msg {
		jobs, err := v.client.ListJobs(context.Background(), v.owner, v.repo, runID)
		return toggleJobsMsg{runID: runID, jobs: jobs, err: err}
	}
}

func (v *ActionsView) onSelectionChange() tea.Cmd {
	switch v.focusPane {
	case 0:
		if wf := v.selectedWorkflow(); wf != nil {
			return v.fetchRuns(wf.ID)
		}
	case 1:
		entry := v.selectedRunTableEntry()
		if entry == nil {
			return nil
		}
		if entry.isRun {
			return v.fetchJobs(entry.runID)
		}
		return v.fetchLogsForInlineJob(entry.runID, entry.jobID)
	case 2:
		if j := v.selectedJob(); j != nil {
			return v.fetchLogs(j.ID)
		}
	}
	return nil
}

func (v *ActionsView) selectedWorkflow() *model.Workflow {
	if v.wfTable.Cursor >= 0 && v.wfTable.Cursor < len(v.workflows) {
		return &v.workflows[v.wfTable.Cursor]
	}
	return nil
}

func (v *ActionsView) selectedRun() *model.Run {
	entry := v.selectedRunTableEntry()
	if entry == nil || !entry.isRun {
		return nil
	}
	for i := range v.runs {
		if v.runs[i].ID == entry.runID {
			return &v.runs[i]
		}
	}
	return nil
}

func (v *ActionsView) selectedJob() *model.Job {
	if v.jobTable.Cursor >= 0 && v.jobTable.Cursor < len(v.jobs) {
		return &v.jobs[v.jobTable.Cursor]
	}
	return nil
}

func (v *ActionsView) fetchWorkflows() tea.Cmd {
	v.loading = true
	v.loadingMsg = "Loading workflows..."
	return func() tea.Msg {
		wfs, err := v.client.ListWorkflows(context.Background(), v.owner, v.repo)
		return actionsMsg{workflows: wfs, err: err}
	}
}

func (v *ActionsView) fetchRuns(workflowID int64) tea.Cmd {
	v.loading = true
	v.loadingMsg = "Loading runs..."
	v.runs = nil
	v.runTable.ClearItems("Loading…")
	v.jobs = nil
	v.jobTable.ClearItems("")
	v.logViewer.SetContent("")
	v.expandedRuns = make(map[int64]bool)
	v.runTableEntries = nil
	return func() tea.Msg {
		runs, err := v.client.ListRuns(context.Background(), v.owner, v.repo, workflowID)
		return runsMsg{runs: runs, err: err}
	}
}

func (v *ActionsView) fetchJobs(runID int64) tea.Cmd {
	v.loading = true
	v.loadingMsg = "Loading jobs..."
	v.jobs = nil
	v.jobTable.ClearItems("Loading…")
	v.logViewer.SetContent("")
	return func() tea.Msg {
		jobs, err := v.client.ListJobs(context.Background(), v.owner, v.repo, runID)
		return jobsMsg{jobs: jobs, err: err}
	}
}

func (v *ActionsView) fetchLogsForInlineJob(runID, jobID int64) tea.Cmd {
	if jobs, ok := v.runJobs[runID]; ok {
		for i := range jobs {
			if jobs[i].ID == jobID && jobs[i].Status != "completed" {
				v.logViewer.SetContent(fmt.Sprintf(
					"  ⏳ Job is %s — logs available after completion.\n\n  Press R to refresh.",
					jobs[i].Status,
				))
				return nil
			}
		}
	}
	return v.fetchLogs(jobID)
}

func (v *ActionsView) fetchLogs(jobID int64) tea.Cmd {
	var targetJob *model.Job
	for i := range v.jobs {
		if v.jobs[i].ID == jobID {
			targetJob = &v.jobs[i]
			break
		}
	}

	if targetJob != nil && targetJob.Status != "completed" {
		v.logViewer.SetContent(fmt.Sprintf(
			"  ⏳ Job is %s — logs available after completion.\n\n  Press R to refresh.",
			targetJob.Status,
		))
		return nil
	}

	v.loading = true
	v.loadingMsg = "Loading logs..."
	v.logViewer.SetContent("")
	return func() tea.Msg {
		logs, err := v.client.GetJobLogs(context.Background(), v.owner, v.repo, jobID)
		return logsMsg{logs: logs, err: err}
	}
}

func (v *ActionsView) rerunSelected() tea.Cmd {
	r := v.selectedRun()
	if r == nil {
		return nil
	}
	v.loading = true
	v.loadingMsg = "Rerunning..."
	runID := r.ID
	return func() tea.Msg {
		err := v.client.RerunWorkflow(context.Background(), v.owner, v.repo, runID)
		return actionResultMsg{action: "rerun", err: err}
	}
}

func (v *ActionsView) rerunFailedSelected() tea.Cmd {
	r := v.selectedRun()
	if r == nil {
		return nil
	}
	v.loading = true
	v.loadingMsg = "Rerunning failed jobs..."
	runID := r.ID
	return func() tea.Msg {
		err := v.client.RerunFailedJobs(context.Background(), v.owner, v.repo, runID)
		return actionResultMsg{action: "rerun-failed", err: err}
	}
}

func (v *ActionsView) cancelSelected() tea.Cmd {
	r := v.selectedRun()
	if r == nil {
		return nil
	}
	v.loading = true
	v.loadingMsg = "Cancelling..."
	runID := r.ID
	return func() tea.Msg {
		err := v.client.CancelRun(context.Background(), v.owner, v.repo, runID)
		return actionResultMsg{action: "cancel", err: err}
	}
}

func (v *ActionsView) triggerSelected() tea.Cmd {
	wf := v.selectedWorkflow()
	if wf == nil {
		return nil
	}
	v.loading = true
	v.loadingMsg = "Triggering workflow..."
	file := filepath.Base(wf.Path)
	return func() tea.Msg {
		err := v.client.TriggerWorkflow(context.Background(), v.owner, v.repo, file, "main", nil)
		return actionResultMsg{action: "trigger", err: err}
	}
}

func (v *ActionsView) refreshCurrent() tea.Cmd {
	return v.fetchWorkflows()
}

func (v *ActionsView) openInBrowser() tea.Cmd {
	var url string
	switch v.focusPane {
	case 1:
		if r := v.selectedRun(); r != nil {
			url = r.HTMLURL
		}
	case 2:
		if j := v.selectedJob(); j != nil {
			url = j.HTMLURL
		}
	default:
		url = fmt.Sprintf("https://github.com/%s/%s/actions", v.owner, v.repo)
	}
	if url == "" {
		return nil
	}
	return openURL(url)
}

func (v *ActionsView) updateWorkflowTable() {
	items := make([]components.TableItem, len(v.workflows))
	for i, wf := range v.workflows {
		name := wf.Name
		if wf.State == "disabled_manually" {
			name += " (disabled)"
		}
		items[i] = components.TableItem{
			ID:      fmt.Sprintf("%d", wf.ID),
			Columns: []string{name},
		}
	}
	v.wfTable.SetItems(items)
}

func (v *ActionsView) updateRunTable() {
	var items []components.TableItem
	v.runTableEntries = nil

	for _, r := range v.runs {
		expanded := v.expandedRuns[r.ID]
		toggle := "▸"
		if expanded {
			toggle = "▾"
		}
		age := shortDuration(time.Since(r.CreatedAt))
		title := truncate(r.Name, 20)
		items = append(items, components.TableItem{
			ID: fmt.Sprintf("run-%d", r.ID),
			Columns: []string{
				toggle,
				r.StatusIcon(),
				fmt.Sprintf("#%d", r.RunNumber),
				title,
				truncate(r.HeadBranch, 14),
				r.Event,
				age,
			},
		})
		v.runTableEntries = append(v.runTableEntries, runTableEntry{isRun: true, runID: r.ID})

		if expanded {
			if jobs, ok := v.runJobs[r.ID]; ok {
				for _, j := range jobs {
					dur := ""
					if !j.StartedAt.IsZero() && !j.CompletedAt.IsZero() {
						dur = shortDuration(j.CompletedAt.Sub(j.StartedAt))
					}
					items = append(items, components.TableItem{
						ID: fmt.Sprintf("job-%d", j.ID),
						Columns: []string{
							" ",
							"  " + j.StatusIcon(),
							j.Name,
							dur,
						},
					})
					v.runTableEntries = append(v.runTableEntries, runTableEntry{isRun: false, runID: r.ID, jobID: j.ID})
				}
			}
		}
	}

	v.runTable.SetItems(items)
}

func (v *ActionsView) updateJobTable() {
	items := make([]components.TableItem, len(v.jobs))
	for i, j := range v.jobs {
		dur := ""
		if !j.StartedAt.IsZero() && !j.CompletedAt.IsZero() {
			dur = shortDuration(j.CompletedAt.Sub(j.StartedAt))
		}
		items[i] = components.TableItem{
			ID:      fmt.Sprintf("%d", j.ID),
			Columns: []string{j.StatusIcon(), j.Name, dur},
		}
	}
	v.jobTable.SetItems(items)
}

func (v *ActionsView) View() string {
	if v.fullscreen {
		v.logViewer.Title = "Logs (f to exit fullscreen)"
		return v.logViewer.View()
	}

	if v.showConfirm {
		return v.renderWithOverlay()
	}

	if v.loading {
		v.logViewer.Title = v.spinner.View() + " " + v.loadingMsg
	} else if v.err != nil {
		v.logViewer.Title = styles.ErrorTxt.Render("Error: " + v.err.Error())
	} else {
		v.logViewer.Title = "Logs"
	}

	filterBar := v.renderFilterBar()
	var panes string
	if v.showLogs {
		panes = lipgloss.JoinHorizontal(lipgloss.Top,
			v.wfTable.View(),
			v.runTable.View(),
			v.jobTable.View(),
			v.logViewer.View(),
		)
	} else {
		panes = lipgloss.JoinHorizontal(lipgloss.Top,
			v.wfTable.View(),
			v.runTable.View(),
			v.jobTable.View(),
		)
	}
	return lipgloss.JoinVertical(lipgloss.Left, filterBar, panes)
}

func (v *ActionsView) renderWithOverlay() string {
	overlay := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Warning).
		Padding(1, 2).
		Render(v.confirmMsg + " (y/n)")

	return lipgloss.Place(v.width, v.height, lipgloss.Center, lipgloss.Center, overlay)
}

func (v *ActionsView) SetFilter(query string) {
	v.filterQuery = query
	v.applyStatusFilter()
}

func (v *ActionsView) applyStatusFilter() {
	sf := statusFilters[v.statusIdx].filter
	lower := strings.ToLower(v.filterQuery)
	var filtered []model.Run
	for _, r := range v.allRuns {
		if sf != "" {
			match := false
			switch sf {
			case "success":
				match = r.Conclusion == "success"
			case "failure":
				match = r.Conclusion == "failure" || r.Conclusion == "timed_out"
			case "in_progress":
				match = r.Status == "in_progress" || r.Status == "queued" || r.Status == "waiting" || r.Status == "pending" || r.Status == "requested"
			}
			if !match {
				continue
			}
		}
		if lower != "" {
			text := strings.ToLower(fmt.Sprintf("%s %s %s", r.Name, r.HeadBranch, r.Conclusion))
			if !strings.Contains(text, lower) {
				continue
			}
		}
		filtered = append(filtered, r)
	}
	if sf == "" && lower == "" {
		v.runs = v.allRuns
	} else {
		v.runs = filtered
	}
	v.expandedRuns = make(map[int64]bool)
	v.runTableEntries = nil
	v.updateRunTable()
}

func (v *ActionsView) renderFilterBar() string {
	sep := lipgloss.NewStyle().Foreground(styles.Muted).Render(" │ ")
	var parts []string
	for i, sf := range statusFilters {
		if i == v.statusIdx {
			parts = append(parts, lipgloss.NewStyle().Bold(true).Foreground(styles.Primary).Render(sf.label))
		} else {
			parts = append(parts, lipgloss.NewStyle().Foreground(styles.Muted).Render(sf.label))
		}
	}
	statusPart := strings.Join(parts, sep)

	wfName := "All Workflows"
	if v.wfTable.Cursor >= 0 && v.wfTable.Cursor < len(v.workflows) {
		wfName = v.workflows[v.wfTable.Cursor].Name
	}
	wfPart := lipgloss.NewStyle().Bold(true).Foreground(styles.Secondary).Render(wfName)

	hint := lipgloss.NewStyle().Foreground(styles.Muted).Italic(true).Render("  s:status")
	line := " " + statusPart + "  ║  " + wfPart + hint

	return lipgloss.NewStyle().Width(v.width).MaxWidth(v.width).Render(line)
}

func truncate(s string, maxLen int) string {
	if ansi.StringWidth(s) <= maxLen {
		return s
	}
	return ansi.Truncate(s, maxLen, "…")
}

func shortDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func openURL(url string) tea.Cmd {
	return func() tea.Msg {
		_ = OpenBrowser(url)
		return nil
	}
}
