package textinput

import (
	"fmt"
	"math/rand"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var cursedRecs = []string{
	"Boku no Pico",
	"Redo Of Healer",
	"Overflow",
	"High School DxD",
	"To-love ru",
}

type Model struct {
	input textinput.Model
}

func New(placeholder string) Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40
	return Model{input: ti}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.input.View()
}

func (m Model) Value() string {
	return m.input.Value()
}

func (m *Model) Focus() {
	m.input.Focus()
}

func (m *Model) RandomPlaceHolder() string {
	return fmt.Sprintf("How about %s?", cursedRecs[rand.Intn(len(cursedRecs))])
}
