package views

import (
	"context"
	"fmt"
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

type notifsMsg struct {
	notifs []model.Notification
	err    error
}

type notifActionMsg struct {
	err error
}

type notifTypeFilter struct {
	label    string
	typeName string // "" = all
}

var notifTypeFilters = []notifTypeFilter{
	{label: "All", typeName: ""},
	{label: "PR", typeName: "PullRequest"},
	{label: "Issue", typeName: "Issue"},
	{label: "Release", typeName: "Release"},
	{label: "CI", typeName: "CheckSuite"},
}

const notifFilterBarH = 1

type NotificationsView struct {
	client *api.Client
	owner  string
	repo   string

	allNotifs []model.Notification
	notifs    []model.Notification
	table     *components.Table
	sidebar   *components.Sidebar

	stateFilter string // "unread" or "all"
	typeIdx     int

	showSidebar bool
	loading     bool
	err         error
	width       int
	height      int
}

func NewNotificationsView(client *api.Client, owner, repo string) *NotificationsView {
	return &NotificationsView{
		client:      client,
		owner:       owner,
		repo:        repo,
		stateFilter: "unread",
		typeIdx:     0,
		table:       components.NewTable("Notifications · Unread · All"),
		sidebar:     components.NewSidebar(),
		showSidebar: true,
	}
}

func (v *NotificationsView) Name() string { return "Notifications" }

func (v *NotificationsView) KeyMap() help.KeyMap { return keys.Notification }

func (v *NotificationsView) Init() tea.Cmd {
	return v.fetchNotifications()
}

func (v *NotificationsView) SetSize(w, h int) {
	v.width = w
	v.height = h
	v.recalcLayout()
}

func (v *NotificationsView) listWidth() int {
	if v.showSidebar {
		return v.width - int(float64(v.width)*0.35)
	}
	return v.width
}

func (v *NotificationsView) recalcLayout() {
	listW := v.width
	sidebarW := 0
	if v.showSidebar {
		sidebarW = int(float64(v.width) * 0.35)
		listW = v.width - sidebarW
	}

	v.table.Width = listW
	v.table.Height = max(4, v.height-notifFilterBarH)
	v.table.Active = true
	v.sidebar.SetSize(sidebarW, v.height)
}

func (v *NotificationsView) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case notifsMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		v.allNotifs = msg.notifs
		v.applyFilter()
		v.updateSidebar()

	case notifActionMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		return v, v.fetchNotifications()

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

func (v *NotificationsView) handleKey(msg tea.KeyMsg) (View, tea.Cmd) {
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
	case key.Matches(msg, keys.Notification.StateToggle):
		if v.stateFilter == "unread" {
			v.stateFilter = "all"
		} else {
			v.stateFilter = "unread"
		}
		v.applyFilter()
		v.updateSidebar()
	case key.Matches(msg, keys.Global.Right):
		v.typeIdx = (v.typeIdx + 1) % len(notifTypeFilters)
		v.applyFilter()
		v.updateSidebar()
	case key.Matches(msg, keys.Global.Left):
		v.typeIdx = (v.typeIdx - 1 + len(notifTypeFilters)) % len(notifTypeFilters)
		v.applyFilter()
		v.updateSidebar()
	case key.Matches(msg, keys.Global.Refresh):
		return v, v.fetchNotifications()
	case key.Matches(msg, keys.Global.Open):
		if n := v.selectedNotif(); n != nil && n.HTMLURL != "" {
			return v, openURL(n.HTMLURL)
		}
	case key.Matches(msg, keys.Notification.MarkRead):
		if n := v.selectedNotif(); n != nil {
			return v, v.markRead(n.ID)
		}
	case key.Matches(msg, keys.Notification.MarkDone):
		if n := v.selectedNotif(); n != nil {
			return v, v.markDone(n.ID)
		}
	case key.Matches(msg, keys.Global.Enter):
		v.showSidebar = !v.showSidebar
		v.recalcLayout()
	}
	return v, nil
}

func (v *NotificationsView) applyFilter() {
	var filtered []model.Notification
	typeFilter := notifTypeFilters[v.typeIdx].typeName
	for _, n := range v.allNotifs {
		if v.stateFilter == "unread" && !n.Unread {
			continue
		}
		if typeFilter != "" && n.Type != typeFilter {
			continue
		}
		filtered = append(filtered, n)
	}
	v.notifs = filtered
	v.updateTitle()
	v.updateTable()
}

func (v *NotificationsView) updateTitle() {
	state := "Unread"
	if v.stateFilter == "all" {
		state = "All"
	}
	v.table.Title = fmt.Sprintf("Notifications · %s · %s", state, notifTypeFilters[v.typeIdx].label)
}

func (v *NotificationsView) selectedNotif() *model.Notification {
	if v.table.Cursor >= 0 && v.table.Cursor < len(v.notifs) {
		return &v.notifs[v.table.Cursor]
	}
	return nil
}

func (v *NotificationsView) fetchNotifications() tea.Cmd {
	v.loading = true
	return func() tea.Msg {
		notifs, err := v.client.ListNotifications(context.Background())
		return notifsMsg{notifs: notifs, err: err}
	}
}

func (v *NotificationsView) markRead(id string) tea.Cmd {
	return func() tea.Msg {
		err := v.client.MarkNotificationRead(context.Background(), id)
		return notifActionMsg{err: err}
	}
}

func (v *NotificationsView) markDone(id string) tea.Cmd {
	return func() tea.Msg {
		err := v.client.MarkNotificationDone(context.Background(), id)
		return notifActionMsg{err: err}
	}
}

func (v *NotificationsView) updateTable() {
	items := make([]components.TableItem, len(v.notifs))
	for i, n := range v.notifs {
		unread := " "
		if n.Unread {
			unread = "●"
		}
		items[i] = components.TableItem{
			ID: n.ID,
			Columns: []string{
				unread,
				n.TypeIcon(),
				truncate(n.Title, 50),
				n.RepoName,
				n.Reason,
			},
		}
	}
	v.table.SetItems(items)
}

func (v *NotificationsView) updateSidebar() {
	n := v.selectedNotif()
	if n == nil {
		v.sidebar.SetContent("No notification selected")
		v.sidebar.Title = "Details"
		return
	}

	v.sidebar.Title = n.Title

	var b strings.Builder
	fmt.Fprintf(&b, "Type:    %s\n", n.Type)
	fmt.Fprintf(&b, "Repo:    %s\n", n.RepoName)
	fmt.Fprintf(&b, "Reason:  %s\n", n.Reason)
	fmt.Fprintf(&b, "Unread:  %v\n", n.Unread)
	fmt.Fprintf(&b, "Updated: %s\n", n.UpdatedAt.Format("2006-01-02 15:04"))

	v.sidebar.SetContent(b.String())
}

func (v *NotificationsView) renderFilterBar() string {
	unreadStyle := lipgloss.NewStyle().Foreground(styles.Muted)
	allStyle := lipgloss.NewStyle().Foreground(styles.Muted)
	if v.stateFilter == "unread" {
		unreadStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.Warning)
	} else {
		allStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.Primary)
	}
	statePart := unreadStyle.Render("Unread") + " " + allStyle.Render("All")

	sep := lipgloss.NewStyle().Foreground(styles.Muted).Render(" │ ")
	var typeParts []string
	for i, tf := range notifTypeFilters {
		if i == v.typeIdx {
			typeParts = append(typeParts, lipgloss.NewStyle().Bold(true).Foreground(styles.Primary).Render(tf.label))
		} else {
			typeParts = append(typeParts, lipgloss.NewStyle().Foreground(styles.Muted).Render(tf.label))
		}
	}
	typePart := strings.Join(typeParts, sep)

	hint := lipgloss.NewStyle().Foreground(styles.Muted).Italic(true).Render("  s:state ←→:type")
	line := " " + statePart + "  ║  " + typePart + hint

	listW := v.listWidth()
	return lipgloss.NewStyle().Width(listW).MaxWidth(listW).Render(line)
}

func (v *NotificationsView) View() string {
	filterBar := v.renderFilterBar()
	listPane := lipgloss.JoinVertical(lipgloss.Left, filterBar, v.table.View())

	if v.showSidebar {
		return lipgloss.JoinHorizontal(lipgloss.Top, listPane, v.sidebar.View())
	}

	return listPane
}
