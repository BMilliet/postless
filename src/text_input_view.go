package src

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TextInputViewModel struct {
	textInput textinput.Model
	title     string
	endValue  *string
	quitting  bool
	styles    Styles
}

func (m TextInputViewModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TextInputViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			*m.endValue = ExitSignal
			m.quitting = true
			return m, tea.Quit
		case "enter":
			*m.endValue = strings.TrimSpace(m.textInput.Value())
			m.quitting = true
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m TextInputViewModel) View() string {
	if m.quitting {
		return ""
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.styles.TitleStyle.Render(m.title),
		m.textInput.View(),
		m.styles.FooterStyle.Render("press enter to confirm • esc to cancel"),
	)
}

func TextFieldView(title, placeHolder string, endValue *string) {
	styles := DefaultStyles()

	ti := textinput.New()
	ti.Placeholder = placeHolder
	ti.SetValue(placeHolder) // Set initial value
	ti.Focus()
	ti.CharLimit = 500 // Increase limit for long JWTs
	ti.Width = 80

	m := TextInputViewModel{
		textInput: ti,
		title:     title,
		endValue:  endValue,
		quitting:  false,
		styles:    *styles,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("TextFieldView -> ", err)
		os.Exit(1)
	}
}
