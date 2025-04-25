package mainmenu

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itsmenewbie03/go-animegg/scrapers/animegg"
	"github.com/itsmenewbie03/go-animegg/ui/textinput"
	"github.com/itsmenewbie03/go-animegg/utils/mpv"
)

var (
	docStyle   = lipgloss.NewStyle().Margin(1, 2)
	controller MainMenuController
	mpvWrapper = mpv.NewMpvWrapper(mpv.ANIMEGG)
)

type item struct {
	ident MenuIdent
	title string
	desc  string
	link  string
}

var delegate = CustomDelegate{
	DefaultDelegate: list.NewDefaultDelegate(),
}

var t = item{
	ident: 1,
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	mode          viewMode
	mainMenuItems list.Model
	searchResult  list.Model
	episodeList   list.Model
	videoList     list.Model
	textInput     textinput.Model
	width         int
	height        int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeMain:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				selected := m.mainMenuItems.SelectedItem()
				if selected != nil {
					i := selected.(item)
					switch i.ident {
					case QUIT:
						return m, tea.Quit
					case UNTRACKED:
						m.mode = modeInput
						m.textInput = textinput.New(m.textInput.RandomPlaceHolder())
						m.textInput.Focus()
						return m, nil
					}
				}
			}
		case tea.WindowSizeMsg:
			// INFO: this only get's fired on the first instance
			h, v := docStyle.GetFrameSize()
			m.width = msg.Width - h
			m.height = msg.Height - v
			m.mainMenuItems.SetSize(msg.Width-h, msg.Height-v)

		}

		var cmd tea.Cmd
		m.mainMenuItems, cmd = m.mainMenuItems.Update(msg)
		return m, cmd

	case modeInput:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				input := m.textInput.Value()
				l, err := controller.Search(input)
				if err != nil {
					panic(err)
				}
				m.mode = modeSearchResult
				m.searchResult = list.New(*l, delegate, 0, 0)
				m.searchResult.Title = "Search Result"
				m.searchResult.SetSize(m.width, m.height)
				return m, nil
			case "esc":
				m.mode = modeMain
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	case modeSearchResult:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				selected := m.searchResult.SelectedItem()
				s := selected.(item)
				pageLink := s.link
				if pageLink == "" {
					panic(fmt.Sprintf("%s has no link available", s.title))
				}
				eps, err := controller.GetEpisodes(pageLink)
				if err != nil {
					panic(err)
				}
				m.mode = modeEpisodeList
				m.episodeList = list.New(*eps, delegate, 0, 0)
				m.episodeList.Title = s.title
				m.episodeList.SetSize(m.width, m.height)
			case "esc":
				m.mode = modeInput
				return m, nil

			}
		case tea.WindowSizeMsg:
			h, v := docStyle.GetFrameSize()
			m.searchResult.SetSize(msg.Width-h, msg.Height-v)
		}

		var cmd tea.Cmd
		m.searchResult, cmd = m.searchResult.Update(msg)
		return m, cmd

	case modeEpisodeList:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				selected := m.episodeList.SelectedItem()
				s := selected.(item)
				episodeLink := s.link
				if episodeLink == "" {
					panic(fmt.Sprintf("%s has no link available", s.title))
				}
				videos, err := controller.GetVideos(episodeLink)
				if err != nil {
					panic(err)
				}
				m.mode = modeVideoList
				m.videoList = list.New(*videos, delegate, 0, 0)
				m.videoList.Title = s.title
				m.videoList.SetSize(m.width, m.height)
			case "esc":
				m.mode = modeSearchResult
				return m, nil
			}
		case tea.WindowSizeMsg:
			h, v := docStyle.GetFrameSize()
			m.episodeList.SetSize(msg.Width-h, msg.Height-v)
		}

		var cmd tea.Cmd
		m.episodeList, cmd = m.episodeList.Update(msg)
		return m, cmd

	case modeVideoList:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				selected := m.videoList.SelectedItem()
				s := selected.(item)
				videoLink := s.link
				err := mpvWrapper.Play(videoLink, s.title)
				if err != nil {
					panic(err)
				}

			case "esc":
				m.mode = modeEpisodeList
				return m, nil
			}
		case tea.WindowSizeMsg:
			h, v := docStyle.GetFrameSize()
			m.videoList.SetSize(msg.Width-h, msg.Height-v)
		}

		var cmd tea.Cmd
		m.videoList, cmd = m.videoList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	switch m.mode {
	case modeInput:
		return docStyle.Render(fmt.Sprintf(
			"🌸 What do you want to watch, senpai?\n\n%s\n\n%s",
			m.textInput.View(),
			"(esc to cancel)",
		))
	case modeSearchResult:
		return docStyle.Render(m.searchResult.View())
	case modeEpisodeList:
		return docStyle.Render(m.episodeList.View())
	case modeVideoList:
		return docStyle.Render(m.videoList.View())
	default: // modeMain or fallback
		return docStyle.Render(m.mainMenuItems.View())
	}
}

func Render(client *animegg.Client) {
	controller = *NewMainMenuController(client)
	items := []list.Item{
		item{ident: CURRENT, title: "Currently Watching", desc: "Animes you're currently watching"},
		item{ident: SHOWALL, title: "Show All", desc: "Display all animes in your list"},
		item{ident: UNTRACKED, title: "Untracked watching", desc: "Animes you're watching but not tracking"},
		item{ident: UPDATE, title: "Update", desc: "Update anime progress or metadata"},
		item{ident: CONTINUE, title: "Continue Last Session", desc: "Resume watching from where you left off"},
		item{ident: QUIT, title: "Quit", desc: "Exit the application"},
	}

	m := model{mainMenuItems: list.New(items, delegate, 0, 0)}
	m.mainMenuItems.Title = "Main Menu"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
