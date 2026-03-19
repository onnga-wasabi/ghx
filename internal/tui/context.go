package tui

import (
	"github.com/onnga-wasabi/ghx/internal/api"
	"github.com/onnga-wasabi/ghx/internal/config"
)

type Context struct {
	Client *api.Client
	Config *config.Config
	Owner  string
	Repo   string
	Width  int
	Height int
}
