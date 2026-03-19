package keys

import "github.com/charmbracelet/bubbles/key"

type NotificationKeyMap struct {
	StateToggle key.Binding
	MarkRead    key.Binding
	MarkDone    key.Binding
	MarkAllRead key.Binding
	Unsubscribe key.Binding
}

var Notification = NotificationKeyMap{
	StateToggle: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "unread/all"),
	),
	MarkRead: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "mark read"),
	),
	MarkDone: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "mark done"),
	),
	MarkAllRead: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("C-r", "mark all read"),
		key.WithDisabled(),
	),
	Unsubscribe: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "unsubscribe"),
		key.WithDisabled(),
	),
}

func (k NotificationKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.StateToggle, k.MarkRead, k.MarkDone}
}

func (k NotificationKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.StateToggle, k.MarkRead, k.MarkDone},
	}
}
