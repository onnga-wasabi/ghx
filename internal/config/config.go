package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PRSections           []Section        `yaml:"prSections"`
	IssueSections        []Section        `yaml:"issuesSections"`
	NotificationSections []Section        `yaml:"notificationsSections"`
	Defaults             Defaults         `yaml:"defaults"`
	Keybindings          KeybindingConfig `yaml:"keybindings"`
	Theme                ThemeConfig      `yaml:"theme"`
}

type Section struct {
	Title   string `yaml:"title"`
	Filters string `yaml:"filters"`
	Limit   int    `yaml:"limit,omitempty"`
}

type Defaults struct {
	PRsLimit            int      `yaml:"prsLimit"`
	IssuesLimit         int      `yaml:"issuesLimit"`
	NotificationsLimit  int      `yaml:"notificationsLimit"`
	View                string   `yaml:"view"`
	Preview             Preview  `yaml:"preview"`
	RefetchIntervalMins int      `yaml:"refetchIntervalMinutes"`
	SmartLayout         *bool    `yaml:"smartLayout,omitempty"`
	Tabs                []string `yaml:"tabs,omitempty"`
}

func (d Defaults) IsSmartLayout() bool {
	if d.SmartLayout == nil {
		return true
	}
	return *d.SmartLayout
}

type Preview struct {
	Open  bool    `yaml:"open"`
	Width float64 `yaml:"width"`
}

type KeybindingConfig struct {
	Universal []KeyBinding `yaml:"universal,omitempty"`
	PRs       []KeyBinding `yaml:"prs,omitempty"`
	Issues    []KeyBinding `yaml:"issues,omitempty"`
	Actions   []KeyBinding `yaml:"actions,omitempty"`
}

type KeyBinding struct {
	Key     string `yaml:"key"`
	Name    string `yaml:"name,omitempty"`
	Builtin string `yaml:"builtin,omitempty"`
	Command string `yaml:"command,omitempty"`
}

type ThemeConfig struct {
	Colors ColorConfig `yaml:"colors,omitempty"`
}

type ColorConfig struct {
	Primary   string `yaml:"primary,omitempty"`
	Secondary string `yaml:"secondary,omitempty"`
	Success   string `yaml:"success,omitempty"`
	Warning   string `yaml:"warning,omitempty"`
	Error     string `yaml:"error,omitempty"`
}

func Load() *Config {
	cfg := defaultConfig()

	paths := configPaths()
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		_ = yaml.Unmarshal(data, cfg)
		break
	}

	return cfg
}

func defaultConfig() *Config {
	return &Config{
		PRSections: []Section{
			{Title: "My Pull Requests", Filters: "is:open author:@me"},
			{Title: "Needs My Review", Filters: "is:open review-requested:@me"},
			{Title: "Involved", Filters: "is:open involves:@me -author:@me"},
			{Title: "Recently Closed", Filters: "is:closed author:@me sort:updated-desc"},
		},
		IssueSections: []Section{
			{Title: "My Issues", Filters: "is:open author:@me"},
			{Title: "Assigned", Filters: "is:open assignee:@me"},
			{Title: "Recently Closed", Filters: "is:closed author:@me sort:updated-desc"},
		},
		NotificationSections: []Section{
			{Title: "All", Filters: ""},
		},
		Defaults: Defaults{
			PRsLimit:            20,
			IssuesLimit:         20,
			NotificationsLimit:  20,
			View:                "prs",
			Preview:             Preview{Open: true, Width: 0.45},
			RefetchIntervalMins: 30,
		},
	}
}

func configPaths() []string {
	var paths []string

	if p := os.Getenv("GHX_CONFIG"); p != "" {
		paths = append(paths, p)
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		paths = append(paths, filepath.Join(xdg, "ghx", "config.yml"))
	}

	home, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(home, ".config", "ghx", "config.yml"))
	}

	return paths
}
