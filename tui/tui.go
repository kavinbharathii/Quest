
package tui

import (
	"strings"

    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/kavinbharathii/quest/index"
)

var (
	titleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("57")).
				Foreground(lipgloss.Color("255")).
				Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
				Padding(0, 1)

	dimStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	freqStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205"))
)

type Model struct {
	input		textinput.Model
	bm25		*index.BM25
	results		[]index.Result
	cursor		int
	chosen		string
	maxHeight	int	
	width		int
	height    	int 
}

func New (b *index.BM25) Model {
	ti := textinput.New()
	ti.Placeholder = "describe the command you're looking for..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 60

	return Model {
		input:		ti,
		bm25:		b,
		maxHeight:	10,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update (msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			m.chosen = ""
			return m, tea.Quit

		case tea.KeyEnter:
			if len(m.results) > 0 {
				m.chosen = m.results[m.cursor].Command
			}
			return m, tea.Quit

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor --
			}
			return m, nil

		case tea.KeyDown:
			if m.cursor < len(m.results) - 1 {
				m.cursor ++
			}
			return m, nil
		}
	
	case tea.WindowSizeMsg:
		m.maxHeight = msg.Height - 6
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}


	var cmd tea.Cmd
	prevValue := m.input.Value()
	m.input, cmd = m.input.Update(msg)

	if m.input.Value() != prevValue {
		query := m.input.Value()
		if query == "" {
			m.results = nil
		} else {
			m.results = m.bm25.Search(query, 10)
		}
		m.cursor = 0
	}

	return m, cmd
}

func (m Model) View() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("⚡️ quest") + "\n")
	sb.WriteString(m.input.View() + "\n\n")

	if len(m.results) == 0 && m.input.Value() != "" {
		sb.WriteString(dimStyle.Render("no results. try different keywords.") + "\n")
	}

	limit := m.maxHeight
	if limit > len(m.results) {
		limit = len(m.results)
	}

	for i, r := range m.results[:limit] {
		line := r.Command

		if i == m.cursor {
			sb.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			sb.WriteString(normalStyle.Render(line) + "\n")
		}
	}

	sb.WriteString("\n" + dimStyle.Render("↑↓ navigate  enter select  esc quit"))

	content := sb.String()
	if m.width > 0 && m.height > 0 {
		box := lipgloss.NewStyle().
			Width(80).
			Render(content)
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			box,
		)
	}
	return content
}

func (m Model) Chosen() string {
	return m.chosen
}
