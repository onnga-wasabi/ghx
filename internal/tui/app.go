package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/onnga-wasabi/ghx/internal/api"
	"github.com/onnga-wasabi/ghx/internal/config"
	"github.com/onnga-wasabi/ghx/internal/tui/components"
	"github.com/onnga-wasabi/ghx/internal/tui/keys"
	"github.com/onnga-wasabi/ghx/internal/tui/styles"
	"github.com/onnga-wasabi/ghx/internal/tui/views"
)

const tabBarHeight = 2

var defaultTabs = []string{"prs", "issues", "actions", "notifications"}

type App struct {
	ctx           *Context
	tabs          *components.Tabs
	views         []views.View
	help          *components.Help
	flash         *components.Flash
	filter        *components.Filter
	customKeys    []customKeyBinding

	width         int
	height        int
	contentHeight int
	ready         bool
}

type customKeyBinding struct {
	binding key.Binding
	command string
}

type customCmdDoneMsg struct{ err error }
type switchTabMsg struct{ tab int }

func NewApp(client *api.Client, owner, repo string) *App {
	cfg := config.Load()
	styles.ApplyTheme(cfg.Theme)

	smart := cfg.Defaults.IsSmartLayout()

	allViews := map[string]views.View{
		"prs":           views.NewPRsView(client, owner, repo, cfg.Defaults.PRsLimit),
		"issues":        views.NewIssuesView(client, owner, repo, cfg.Defaults.IssuesLimit),
		"actions":       views.NewActionsView(client, owner, repo, smart),
		"notifications": views.NewNotificationsView(client, owner, repo),
	}

	tabConfig := cfg.Defaults.Tabs
	if len(tabConfig) == 0 {
		tabConfig = defaultTabs
	}

	var viewList []views.View
	for _, name := range tabConfig {
		if v, ok := allViews[name]; ok {
			viewList = append(viewList, v)
		}
	}
	if len(viewList) == 0 {
		for _, name := range defaultTabs {
			viewList = append(viewList, allViews[name])
		}
	}

	// Wire cross-view navigation: PRs → Actions
	if prsView, ok := allViews["prs"].(*views.PRsView); ok {
		if actionsView, ok := allViews["actions"].(*views.ActionsView); ok {
			actionsIdx := -1
			for i, v := range viewList {
				if v.Name() == "Actions" {
					actionsIdx = i
					break
				}
			}
			if actionsIdx >= 0 {
				prsView.SetOnNavigateToActions(func(branch string) tea.Cmd {
					return tea.Batch(
						func() tea.Msg { return switchTabMsg{tab: actionsIdx} },
						actionsView.ShowRunsForBranch(branch),
					)
				})
			}
		}
	}

	tabNames := make([]string, len(viewList))
	for i, v := range viewList {
		tabNames[i] = v.Name()
	}

	activeTab := 0
	for i, v := range viewList {
		if v.Name() == viewNameFromConfig(cfg.Defaults.View) {
			activeTab = i
			break
		}
	}

	tabs := components.NewTabs(tabNames)
	tabs.SetActive(activeTab)

	customKeys := buildCustomKeys(cfg.Keybindings.Universal)

	return &App{
		ctx: &Context{
			Client: client,
			Config: cfg,
			Owner:  owner,
			Repo:   repo,
		},
		tabs:       tabs,
		views:      viewList,
		help:       components.NewHelp(),
		flash:      components.NewFlash(),
		filter:     components.NewFilter(),
		customKeys: customKeys,
	}
}

func viewNameFromConfig(cfgView string) string {
	switch cfgView {
	case "prs":
		return "PRs"
	case "issues":
		return "Issues"
	case "actions":
		return "Actions"
	case "notifications":
		return "Notifications"
	default:
		return "PRs"
	}
}

func buildCustomKeys(bindings []config.KeyBinding) []customKeyBinding {
	var result []customKeyBinding
	for _, kb := range bindings {
		if kb.Command == "" || kb.Key == "" {
			continue
		}
		result = append(result, customKeyBinding{
			binding: key.NewBinding(
				key.WithKeys(kb.Key),
				key.WithHelp(kb.Key, kb.Name),
			),
			command: kb.Command,
		})
	}
	return result
}

func (a *App) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, v := range a.views {
		cmds = append(cmds, v.Init())
	}
	return tea.Batch(cmds...)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
		a.recalcSizes()
		return a, nil

	case components.FlashClearMsg:
		a.flash.Clear()
		return a, nil

	case switchTabMsg:
		a.tabs.SetActive(msg.tab)
		a.recalcSizes()
		return a, nil

	case tea.KeyMsg:
		if a.help.IsShowingAll() {
			switch {
			case key.Matches(msg, keys.Global.Quit), key.Matches(msg, keys.Global.Help):
				a.help.Toggle()
				return a, nil
			}
			return a, nil
		}

		if a.filter.IsActive() {
			var cmd tea.Cmd
			a.filter, cmd = a.filter.Update(msg)
			if !a.filter.IsActive() {
				a.views[a.tabs.Active].SetFilter(a.filter.Value())
			}
			return a, cmd
		}

		if im, ok := a.views[a.tabs.Active].(views.InputModeView); ok && im.IsInputMode() {
			activeView := a.views[a.tabs.Active]
			var newView views.View
			var cmd tea.Cmd
			newView, cmd = activeView.Update(msg)
			a.views[a.tabs.Active] = newView
			return a, cmd
		}

		switch {
		case key.Matches(msg, keys.Global.Quit):
			return a, tea.Quit

		case key.Matches(msg, keys.Global.Help):
			a.help.Toggle()
			return a, nil

		case key.Matches(msg, keys.Global.NextTab):
			a.tabs.Next()
			a.recalcSizes()
			return a, nil

		case key.Matches(msg, keys.Global.PrevTab):
			a.tabs.Prev()
			a.recalcSizes()
			return a, nil
		}

		// Dynamic tab number keys (1-9)
		for i := 0; i < len(a.views) && i < 9; i++ {
			if msg.String() == fmt.Sprintf("%d", i+1) {
				a.tabs.SetActive(i)
				a.recalcSizes()
				return a, nil
			}
		}

		if key.Matches(msg, keys.Global.Filter) {
			if interceptor, ok := a.views[a.tabs.Active].(views.FilterKeyInterceptor); ok && interceptor.WantsFilterKey() {
				// Let the view handle / (e.g., log search)
			} else {
				return a, a.filter.Activate()
			}
		}

		for _, ck := range a.customKeys {
			if key.Matches(msg, ck.binding) {
				cmd := ck.command
				return a, tea.ExecProcess(exec.Command("sh", "-c", cmd), func(err error) tea.Msg {
					return customCmdDoneMsg{err: err}
				})
			}
		}
	}

	if _, ok := msg.(customCmdDoneMsg); ok {
		return a, nil
	}

	if _, isKey := msg.(tea.KeyMsg); isKey {
		activeView := a.views[a.tabs.Active]
		var newView views.View
		var cmd tea.Cmd
		newView, cmd = activeView.Update(msg)
		a.views[a.tabs.Active] = newView
		cmds = append(cmds, cmd)
	} else {
		for i, v := range a.views {
			var newView views.View
			var cmd tea.Cmd
			newView, cmd = v.Update(msg)
			a.views[i] = newView
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *App) View() string {
	if !a.ready {
		return "Loading..."
	}

	a.tabs.Width = a.width
	tabBar := a.tabs.View()
	footer := a.renderFooter()

	activeView := a.views[a.tabs.Active]
	rawContent := activeView.View()

	content := lipgloss.NewStyle().
		Width(a.width).
		Height(a.contentHeight).
		MaxHeight(a.contentHeight).
		Render(rawContent)

	base := lipgloss.JoinVertical(lipgloss.Left, tabBar, content, footer)

	if a.help.IsShowingAll() {
		overlayW := min(a.width-4, 100)
		a.help.SetWidth(overlayW - 8)
		helpContent := a.help.FullView(keys.Global, activeView.KeyMap())
		a.help.SetWidth(a.width)

		title := lipgloss.NewStyle().Bold(true).Foreground(styles.Primary).Render("  Keyboard Shortcuts  ")
		dismiss := lipgloss.NewStyle().Foreground(styles.Muted).Italic(true).Render("Press ? or q to close")

		inner := title + "\n\n" + helpContent + "\n\n" + dismiss
		overlay := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.Primary).
			Background(styles.BgOverlay).
			Padding(1, 3).
			Render(inner)

		return compositeOverlay(base, overlay, a.width, a.height)
	}

	return base
}

func (a *App) renderFooter() string {
	activeView := a.views[a.tabs.Active]
	helpText := a.help.View(keys.Global, activeView.KeyMap())

	flashText := a.flash.View()
	filterText := a.filter.View()

	left := helpText
	right := ""
	if filterText != "" {
		right = filterText
	}
	if flashText != "" {
		right = flashText
	}

	repoInfo := lipgloss.NewStyle().Foreground(styles.Muted).Render(a.ctx.Owner + "/" + a.ctx.Repo)

	w := a.width
	used := lipgloss.Width(left) + lipgloss.Width(right) + lipgloss.Width(repoInfo)
	gap := max(1, w-used)
	spacer := lipgloss.NewStyle().Width(gap).Render("")

	if right != "" {
		return lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, right)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, repoInfo)
}

func (a *App) recalcSizes() {
	a.help.SetWidth(a.width)
	a.filter.SetWidth(a.width)

	footer := a.renderFooter()
	footerH := lipgloss.Height(footer)
	a.contentHeight = max(1, a.height-tabBarHeight-footerH)

	for _, v := range a.views {
		v.SetSize(a.width, a.contentHeight)
	}
}

// compositeOverlay renders the overlay centered on top of the background,
// preserving background content on either side and above/below.
func compositeOverlay(bg, fg string, totalW, totalH int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	fgH := len(fgLines)
	fgW := lipgloss.Width(fg)

	startY := max(0, (totalH-fgH)/2)
	startX := max(0, (totalW-fgW)/2)

	for len(bgLines) < totalH {
		bgLines = append(bgLines, strings.Repeat(" ", totalW))
	}

	for i, fgLine := range fgLines {
		y := startY + i
		if y >= len(bgLines) {
			break
		}
		bgLine := bgLines[y]
		lineW := ansi.StringWidth(bgLine)
		if lineW < totalW {
			bgLine += strings.Repeat(" ", totalW-lineW)
		}

		left := ansi.Truncate(bgLine, startX, "")
		leftW := ansi.StringWidth(left)
		if leftW < startX {
			left += strings.Repeat(" ", startX-leftW)
		}

		right := ""
		rightStart := startX + fgW
		if rightStart < totalW {
			right = ansi.Cut(bgLine, rightStart, totalW)
		}

		bgLines[y] = left + fgLine + right
	}

	if len(bgLines) > totalH {
		bgLines = bgLines[:totalH]
	}

	return strings.Join(bgLines, "\n")
}
