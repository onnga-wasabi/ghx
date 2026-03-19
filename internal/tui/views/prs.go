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

type prsMsg struct {
	prs []model.PR
	err error
}

type prActionMsg struct {
	action string
	err    error
}

type prScope struct {
	label  string
	filter string
}

var prScopes = []prScope{
	{label: "All", filter: ""},
	{label: "Mine", filter: "author:@me"},
	{label: "Review", filter: "review-requested:@me"},
	{label: "Involved", filter: "involves:@me -author:@me"},
}

const prFilterBarH = 1

type PRsView struct {
	client *api.Client
	owner  string
	repo   string
	limit  int

	stateFilter string // "open" or "closed"
	scopeIdx    int

	prs     []model.PR
	table   *components.Table
	sidebar *components.Sidebar

	showSidebar bool

	loading bool
	err     error
	width   int
	height  int

	onNavigateToActions func(branch string) tea.Cmd
}

func NewPRsView(client *api.Client, owner, repo string, limit int) *PRsView {
	return &PRsView{
		client:      client,
		owner:       owner,
		repo:        repo,
		limit:       limit,
		stateFilter: "open",
		scopeIdx:    0,
		table:       components.NewTable("Pull Requests · Open · All"),
		sidebar:     components.NewSidebar(),
		showSidebar: true,
	}
}

func (v *PRsView) SetOnNavigateToActions(fn func(branch string) tea.Cmd) {
	v.onNavigateToActions = fn
}

func (v *PRsView) Name() string { return "PRs" }

func (v *PRsView) KeyMap() help.KeyMap { return keys.PR }

func (v *PRsView) Init() tea.Cmd {
	return v.fetchPRs()
}

func (v *PRsView) SetSize(w, h int) {
	v.width = w
	v.height = h
	v.recalcLayout()
}

func (v *PRsView) listWidth() int {
	if v.showSidebar {
		return v.width - int(float64(v.width)*0.45)
	}
	return v.width
}

func (v *PRsView) recalcLayout() {
	listW := v.width
	sidebarW := 0
	if v.showSidebar {
		sidebarW = int(float64(v.width) * 0.45)
		listW = v.width - sidebarW
	}

	v.table.Width = listW
	v.table.Height = max(4, v.height-prFilterBarH)
	v.table.Active = true

	v.sidebar.SetSize(sidebarW, v.height)
}

func (v *PRsView) buildQuery() string {
	q := fmt.Sprintf("is:pr repo:%s/%s is:%s", v.owner, v.repo, v.stateFilter)
	if scope := prScopes[v.scopeIdx]; scope.filter != "" {
		q += " " + scope.filter
	}
	return q
}

func (v *PRsView) fetchPRs() tea.Cmd {
	v.loading = true
	v.table.ClearItems("Loading…")
	query := v.buildQuery()
	limit := v.limit
	return func() tea.Msg {
		prs, err := v.client.SearchPRs(context.Background(), query, limit)
		return prsMsg{prs: prs, err: err}
	}
}

func (v *PRsView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case prsMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.prs = msg.prs
		v.updateTable()
		v.updateSidebar()

	case prActionMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
		}
		return v, v.fetchPRs()

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

func (v *PRsView) handleKey(msg tea.KeyMsg) (View, tea.Cmd) {
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
	case key.Matches(msg, keys.PR.StateToggle):
		if v.stateFilter == "open" {
			v.stateFilter = "closed"
		} else {
			v.stateFilter = "open"
		}
		v.updateTitle()
		return v, v.fetchPRs()
	case key.Matches(msg, keys.Global.Right):
		v.scopeIdx = (v.scopeIdx + 1) % len(prScopes)
		v.updateTitle()
		return v, v.fetchPRs()
	case key.Matches(msg, keys.Global.Left):
		v.scopeIdx = (v.scopeIdx - 1 + len(prScopes)) % len(prScopes)
		v.updateTitle()
		return v, v.fetchPRs()
	case key.Matches(msg, keys.Global.Refresh):
		return v, v.fetchPRs()
	case key.Matches(msg, keys.Global.Open):
		if pr := v.selectedPR(); pr != nil {
			return v, openURL(pr.URL)
		}
	case key.Matches(msg, keys.PR.Enhance):
		if pr := v.selectedPR(); pr != nil && v.onNavigateToActions != nil && pr.HeadRef != "" {
			return v, v.onNavigateToActions(pr.HeadRef)
		}
	case key.Matches(msg, keys.PR.Diff):
		if pr := v.selectedPR(); pr != nil {
			return v, openURL(pr.URL + "/files")
		}
	case key.Matches(msg, keys.PR.Approve):
		return v, v.approvePR()
	case key.Matches(msg, keys.PR.Merge):
		return v, v.mergePR()
	case key.Matches(msg, keys.PR.Close):
		return v, v.closePR()
	case key.Matches(msg, keys.PR.Comment):
		return v, v.commentPR()
	case key.Matches(msg, keys.Global.Enter):
		v.showSidebar = !v.showSidebar
		v.recalcLayout()
	}
	return v, nil
}

func (v *PRsView) updateTitle() {
	state := "Open"
	if v.stateFilter == "closed" {
		state = "Closed"
	}
	v.table.Title = fmt.Sprintf("Pull Requests · %s · %s", state, prScopes[v.scopeIdx].label)
}

func (v *PRsView) prOwnerRepo() (string, string) {
	if pr := v.selectedPR(); pr != nil {
		parts := strings.SplitN(pr.RepoName, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return v.owner, v.repo
}

func (v *PRsView) approvePR() tea.Cmd {
	pr := v.selectedPR()
	if pr == nil {
		return nil
	}
	owner, repo := v.prOwnerRepo()
	num := pr.Number
	return func() tea.Msg {
		err := v.client.ApprovePR(context.Background(), owner, repo, num)
		return prActionMsg{action: "approve", err: err}
	}
}

func (v *PRsView) mergePR() tea.Cmd {
	pr := v.selectedPR()
	if pr == nil {
		return nil
	}
	owner, repo := v.prOwnerRepo()
	num := pr.Number
	return func() tea.Msg {
		err := v.client.MergePR(context.Background(), owner, repo, num)
		return prActionMsg{action: "merge", err: err}
	}
}

func (v *PRsView) closePR() tea.Cmd {
	pr := v.selectedPR()
	if pr == nil {
		return nil
	}
	owner, repo := v.prOwnerRepo()
	num := pr.Number
	return func() tea.Msg {
		err := v.client.ClosePR(context.Background(), owner, repo, num)
		return prActionMsg{action: "close", err: err}
	}
}

func (v *PRsView) commentPR() tea.Cmd {
	pr := v.selectedPR()
	if pr == nil {
		return nil
	}
	owner, repo := v.prOwnerRepo()
	num := fmt.Sprintf("%d", pr.Number)
	c := exec.Command("gh", "pr", "comment", num, "-R", owner+"/"+repo)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return prActionMsg{action: "comment", err: err}
	})
}

func (v *PRsView) selectedPR() *model.PR {
	if v.table.Cursor >= 0 && v.table.Cursor < len(v.prs) {
		return &v.prs[v.table.Cursor]
	}
	return nil
}

func (v *PRsView) updateTable() {
	items := make([]components.TableItem, len(v.prs))
	for i, pr := range v.prs {
		pass, fail, pending := pr.ChecksSummary()
		checkStr := ""
		if pass+fail+pending > 0 {
			checkStr = fmt.Sprintf("%d/%d", pass, pass+fail+pending)
		}
		items[i] = components.TableItem{
			ID: fmt.Sprintf("%d", pr.Number),
			Columns: []string{
				pr.StatusIcon(),
				fmt.Sprintf("#%d", pr.Number),
				truncate(pr.Title, 40),
				pr.Author,
				checkStr,
			},
		}
	}
	v.table.SetItems(items)
}

func (v *PRsView) updateSidebar() {
	pr := v.selectedPR()
	if pr == nil {
		v.sidebar.SetContent("No PR selected")
		v.sidebar.Title = "Preview"
		return
	}

	v.sidebar.Title = fmt.Sprintf("#%d %s", pr.Number, pr.Title)

	var b strings.Builder
	fmt.Fprintf(&b, "Author:  %s\n", pr.Author)
	fmt.Fprintf(&b, "Branch:  %s → %s\n", pr.HeadRef, pr.BaseRef)
	fmt.Fprintf(&b, "State:   %s\n", pr.State)
	fmt.Fprintf(&b, "Lines:   +%d -%d\n", pr.Additions, pr.Deletions)
	if len(pr.Labels) > 0 {
		fmt.Fprintf(&b, "Labels:  %s\n", strings.Join(pr.Labels, ", "))
	}
	if pr.ReviewState != "" {
		fmt.Fprintf(&b, "Review:  %s\n", pr.ReviewState)
	}

	pass, fail, pending := pr.ChecksSummary()
	if pass+fail+pending > 0 {
		b.WriteString("\n── Checks ──\n")
		for _, c := range pr.Checks {
			icon := "⏳"
			if c.Conclusion == "SUCCESS" {
				icon = "✓"
			} else if c.Conclusion == "FAILURE" {
				icon = "✗"
			}
			fmt.Fprintf(&b, "  %s %s\n", icon, c.Name)
		}
	}

	if pr.Body != "" {
		b.WriteString("\n── Description ──\n")
		b.WriteString(pr.Body)
	}

	v.sidebar.SetContent(b.String())
}

func (v *PRsView) renderFilterBar() string {
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
	for i, s := range prScopes {
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

func (v *PRsView) View() string {
	filterBar := v.renderFilterBar()
	listPane := lipgloss.JoinVertical(lipgloss.Left, filterBar, v.table.View())

	if v.showSidebar {
		return lipgloss.JoinHorizontal(lipgloss.Top, listPane, v.sidebar.View())
	}

	return listPane
}
