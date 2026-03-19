package views

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/onnga-wasabi/ghx/internal/api"
	"github.com/onnga-wasabi/ghx/internal/model"
	"github.com/onnga-wasabi/ghx/internal/tui/components"
	"github.com/onnga-wasabi/ghx/internal/tui/keys"
	"github.com/onnga-wasabi/ghx/internal/tui/styles"
)

type issuesMsg struct {
	issues []model.Issue
	err    error
}

type issueActionMsg struct {
	action string
	err    error
}

type issueScope struct {
	label  string
	filter string
}

var issueScopes = []issueScope{
	{label: "All", filter: ""},
	{label: "Mine", filter: "author:@me"},
	{label: "Assigned", filter: "assignee:@me"},
	{label: "Mentioned", filter: "mentions:@me"},
}

const issueFilterBarH = 1

type IssuesView struct {
	client *api.Client
	owner  string
	repo   string
	limit  int

	stateFilter string // "open" or "closed"
	scopeIdx    int
	filterQuery string

	allIssues []model.Issue
	issues    []model.Issue
	table   *components.Table
	sidebar *components.Sidebar

	showSidebar bool

	loading bool
	err     error
	width   int
	height  int
}

func NewIssuesView(client *api.Client, owner, repo string, limit int) *IssuesView {
	return &IssuesView{
		client:      client,
		owner:       owner,
		repo:        repo,
		limit:       limit,
		stateFilter: "open",
		scopeIdx:    0,
		table:       components.NewTable("Issues · Open · All"),
		sidebar:     components.NewSidebar(),
		showSidebar: true,
	}
}

func (v *IssuesView) Name() string { return "Issues" }

func (v *IssuesView) KeyMap() help.KeyMap { return keys.Issue }

func (v *IssuesView) Init() tea.Cmd {
	return v.fetchIssues()
}

func (v *IssuesView) SetSize(w, h int) {
	v.width = w
	v.height = h
	v.recalcLayout()
}

func (v *IssuesView) listWidth() int {
	if v.showSidebar {
		return v.width - int(float64(v.width)*0.45)
	}
	return v.width
}

func (v *IssuesView) recalcLayout() {
	listW := v.width
	sidebarW := 0
	if v.showSidebar {
		sidebarW = int(float64(v.width) * 0.45)
		listW = v.width - sidebarW
	}

	v.table.Width = listW
	v.table.Height = max(4, v.height-issueFilterBarH)
	v.table.Active = true

	v.sidebar.SetSize(sidebarW, v.height)
}

func (v *IssuesView) buildQuery() string {
	q := fmt.Sprintf("is:issue repo:%s/%s is:%s", v.owner, v.repo, v.stateFilter)
	if scope := issueScopes[v.scopeIdx]; scope.filter != "" {
		q += " " + scope.filter
	}
	return q
}

func (v *IssuesView) fetchIssues() tea.Cmd {
	v.loading = true
	v.table.ClearItems("Loading…")
	query := v.buildQuery()
	limit := v.limit
	return func() tea.Msg {
		issues, err := v.client.SearchIssues(context.Background(), query, limit)
		return issuesMsg{issues: issues, err: err}
	}
}

func (v *IssuesView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case issuesMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.allIssues = msg.issues
		v.applyFilter()
		v.updateSidebar()

	case issueActionMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
		}
		return v, v.fetchIssues()

	case tea.KeyMsg:
		return v.handleKey(msg)
	}

	if v.showSidebar {
		var cmd tea.Cmd
		v.sidebar, cmd = v.sidebar.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v *IssuesView) handleKey(msg tea.KeyMsg) (View, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Global.Up):
		v.table.MoveUp()
		v.updateSidebar()
	case key.Matches(msg, keys.Global.Down):
		v.table.MoveDown()
		v.updateSidebar()
	case key.Matches(msg, keys.Global.FirstLine):
		v.table.GoToFirst()
		v.updateSidebar()
	case key.Matches(msg, keys.Global.LastLine):
		v.table.GoToLast()
		v.updateSidebar()
	case key.Matches(msg, keys.Issue.StateToggle):
		if v.stateFilter == "open" {
			v.stateFilter = "closed"
		} else {
			v.stateFilter = "open"
		}
		v.updateTitle()
		return v, v.fetchIssues()
	case key.Matches(msg, keys.Global.Right):
		v.scopeIdx = (v.scopeIdx + 1) % len(issueScopes)
		v.updateTitle()
		return v, v.fetchIssues()
	case key.Matches(msg, keys.Global.Left):
		v.scopeIdx = (v.scopeIdx - 1 + len(issueScopes)) % len(issueScopes)
		v.updateTitle()
		return v, v.fetchIssues()
	case key.Matches(msg, keys.Global.Refresh):
		return v, v.fetchIssues()
	case key.Matches(msg, keys.Global.Open):
		if issue := v.selectedIssue(); issue != nil {
			return v, openURL(issue.URL)
		}
	case key.Matches(msg, keys.Issue.Close):
		return v, v.closeIssue()
	case key.Matches(msg, keys.Issue.Reopen):
		return v, v.reopenIssue()
	case key.Matches(msg, keys.Issue.Comment):
		return v, v.commentIssue()
	case key.Matches(msg, keys.Global.Enter):
		v.showSidebar = !v.showSidebar
		v.recalcLayout()
	}
	return v, nil
}

func (v *IssuesView) updateTitle() {
	state := "Open"
	if v.stateFilter == "closed" {
		state = "Closed"
	}
	v.table.Title = fmt.Sprintf("Issues · %s · %s", state, issueScopes[v.scopeIdx].label)
}

func (v *IssuesView) issueOwnerRepo() (string, string) {
	if issue := v.selectedIssue(); issue != nil {
		parts := strings.SplitN(issue.RepoName, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return v.owner, v.repo
}

func (v *IssuesView) closeIssue() tea.Cmd {
	issue := v.selectedIssue()
	if issue == nil {
		return nil
	}
	owner, repo := v.issueOwnerRepo()
	num := issue.Number
	return func() tea.Msg {
		err := v.client.CloseIssue(context.Background(), owner, repo, num)
		return issueActionMsg{action: "close", err: err}
	}
}

func (v *IssuesView) reopenIssue() tea.Cmd {
	issue := v.selectedIssue()
	if issue == nil {
		return nil
	}
	owner, repo := v.issueOwnerRepo()
	num := issue.Number
	return func() tea.Msg {
		err := v.client.ReopenIssue(context.Background(), owner, repo, num)
		return issueActionMsg{action: "reopen", err: err}
	}
}

func (v *IssuesView) commentIssue() tea.Cmd {
	issue := v.selectedIssue()
	if issue == nil {
		return nil
	}
	owner, repo := v.issueOwnerRepo()
	num := fmt.Sprintf("%d", issue.Number)
	c := exec.Command("gh", "issue", "comment", num, "-R", owner+"/"+repo)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return issueActionMsg{action: "comment", err: err}
	})
}

func (v *IssuesView) selectedIssue() *model.Issue {
	if v.table.Cursor >= 0 && v.table.Cursor < len(v.issues) {
		return &v.issues[v.table.Cursor]
	}
	return nil
}

func (v *IssuesView) SetFilter(query string) {
	v.filterQuery = query
	v.applyFilter()
}

func (v *IssuesView) applyFilter() {
	if v.filterQuery == "" {
		v.issues = v.allIssues
	} else {
		lower := strings.ToLower(v.filterQuery)
		var filtered []model.Issue
		for _, issue := range v.allIssues {
			text := strings.ToLower(fmt.Sprintf("#%d %s %s", issue.Number, issue.Title, issue.Author))
			if strings.Contains(text, lower) {
				filtered = append(filtered, issue)
			}
		}
		v.issues = filtered
	}
	v.updateTable()
}

func (v *IssuesView) updateTable() {
	items := make([]components.TableItem, len(v.issues))
	for i, issue := range v.issues {
		items[i] = components.TableItem{
			ID: fmt.Sprintf("%d", issue.Number),
			Columns: []string{
				issue.StatusIcon(),
				fmt.Sprintf("#%d", issue.Number),
				truncate(issue.Title, 40),
				issue.Author,
				fmt.Sprintf("💬%d", issue.Comments),
			},
		}
	}
	v.table.SetItems(items)
}

func (v *IssuesView) updateSidebar() {
	issue := v.selectedIssue()
	if issue == nil {
		v.sidebar.SetContent("No issue selected")
		v.sidebar.Title = "Preview"
		return
	}

	v.sidebar.Title = fmt.Sprintf("#%d %s", issue.Number, issue.Title)

	var b strings.Builder
	fmt.Fprintf(&b, "Author:    %s\n", issue.Author)
	fmt.Fprintf(&b, "State:     %s\n", issue.State)
	fmt.Fprintf(&b, "Comments:  %d\n", issue.Comments)
	if len(issue.Labels) > 0 {
		fmt.Fprintf(&b, "Labels:    %s\n", strings.Join(issue.Labels, ", "))
	}
	if len(issue.Assignees) > 0 {
		fmt.Fprintf(&b, "Assignees: %s\n", strings.Join(issue.Assignees, ", "))
	}

	if issue.Body != "" {
		b.WriteString("\n── Description ──\n")
		b.WriteString(issue.Body)
	}

	v.sidebar.SetContent(b.String())
}

func (v *IssuesView) renderFilterBar() string {
	openStyle := lipgloss.NewStyle().Foreground(styles.Muted)
	closedStyle := lipgloss.NewStyle().Foreground(styles.Muted)
	if v.stateFilter == "open" {
		openStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.Success)
	} else {
		closedStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.Secondary)
	}
	statePart := openStyle.Render("Open") + " " + closedStyle.Render("Closed")

	sep := lipgloss.NewStyle().Foreground(styles.Muted).Render(" │ ")
	var scopeParts []string
	for i, s := range issueScopes {
		if i == v.scopeIdx {
			scopeParts = append(scopeParts, lipgloss.NewStyle().Bold(true).Foreground(styles.Primary).Render(s.label))
		} else {
			scopeParts = append(scopeParts, lipgloss.NewStyle().Foreground(styles.Muted).Render(s.label))
		}
	}
	scopePart := strings.Join(scopeParts, sep)

	hint := lipgloss.NewStyle().Foreground(styles.Muted).Italic(true).Render("  s:state ←→:scope")
	line := " " + statePart + "  ║  " + scopePart + hint

	listW := v.listWidth()
	return lipgloss.NewStyle().Width(listW).MaxWidth(listW).Render(line)
}

func (v *IssuesView) View() string {
	filterBar := v.renderFilterBar()
	listPane := lipgloss.JoinVertical(lipgloss.Left, filterBar, v.table.View())

	if v.showSidebar {
		return lipgloss.JoinHorizontal(lipgloss.Top, listPane, v.sidebar.View())
	}

	return listPane
}
