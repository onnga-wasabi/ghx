package keys

import "github.com/charmbracelet/bubbles/key"

type ActionsKeyMap struct {
	StatusToggle key.Binding
	Trigger      key.Binding
	Rerun        key.Binding
	RerunFailed  key.Binding
	Cancel       key.Binding
	Fullscreen   key.Binding
	LogToggle    key.Binding
	WordWrap     key.Binding
	SearchNext   key.Binding
	SearchPrev   key.Binding
	ExpandAll    key.Binding
	CollapseAll  key.Binding
	NextSection  key.Binding
	PrevSection  key.Binding
}

var Actions = ActionsKeyMap{
	StatusToggle: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "status filter"),
	),
	Trigger: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "trigger workflow"),
	),
	Rerun: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "rerun"),
	),
	RerunFailed: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("C-r", "rerun failed"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "cancel"),
	),
	Fullscreen: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "fullscreen log"),
	),
	LogToggle: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "toggle log pane"),
	),
	WordWrap: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "word wrap"),
	),
	SearchNext: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "next match"),
	),
	SearchPrev: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("N", "prev match"),
	),
	ExpandAll: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "expand all sections"),
	),
	CollapseAll: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "collapse all sections"),
	),
	NextSection: key.NewBinding(
		key.WithKeys("]"),
		key.WithHelp("]", "next section"),
	),
	PrevSection: key.NewBinding(
		key.WithKeys("["),
		key.WithHelp("[", "prev section"),
	),
}

func (k ActionsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.StatusToggle, k.Rerun, k.Cancel, k.Fullscreen, k.ExpandAll, k.CollapseAll}
}

func (k ActionsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.StatusToggle, k.Trigger, k.Rerun, k.RerunFailed, k.Cancel},
		{k.Fullscreen, k.LogToggle, k.WordWrap, k.SearchNext, k.SearchPrev},
		{k.ExpandAll, k.CollapseAll, k.NextSection, k.PrevSection},
	}
}
