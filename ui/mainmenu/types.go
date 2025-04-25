package mainmenu

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type MenuIdent int

const (
	CURRENT MenuIdent = iota
	SHOWALL
	UNTRACKED
	UPDATE
	CONTINUE
	QUIT
)

type viewMode int

const (
	modeMain viewMode = iota
	modeInput
	modeSearchResult
	modeEpisodeList
	modeVideoList
)

type CustomDelegate struct {
	list.DefaultDelegate
}

func (d CustomDelegate) ShortHelp() []key.Binding {
	return append(d.DefaultDelegate.ShortHelp(), key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "previous screen"),
	))
}

func (d CustomDelegate) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		d.ShortHelp(),
	}
}
