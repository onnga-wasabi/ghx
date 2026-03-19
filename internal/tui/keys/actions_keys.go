package keys

import "github.com/charmbracelet/bubbles/key"

type ActionsKeyMap struct {
	Trigger      key.Binding
	Rerun        key.Binding
	RerunFailed  key.Binding
	Cancel       key.Binding
	Fullscreen   key.Binding
	WordWrap     key.Binding
	SearchNext   key.Binding
	SearchPrev   key.Binding
}

var Actions = ActionsKeyMap{
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
}

func (k ActionsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Trigger, k.Rerun, k.Cancel, k.Fullscreen}
}

func (k ActionsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Trigger, k.Rerun, k.RerunFailed, k.Cancel},
		{k.Fullscreen, k.WordWrap, k.SearchNext, k.SearchPrev},
	}
}
