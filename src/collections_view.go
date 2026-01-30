package src

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CollectionsViewModel struct {
	collections   []Collection
	config        *ConfigJSON
	secret        *SecretJSON
	configLoader  *ConfigLoader
	fileManager   FileManagerInterface
	currentPage   int
	cursor        int
	viewportStart int
	maxVisible    int
	selected      *string
	quitting      bool
	styles        *Styles
	searchMode    bool
	searchQuery   string
	filteredList  []RequestItem
	totalPages    int // Collections + Settings page
}

func NewCollectionsViewModel(collections []Collection, config *ConfigJSON, secret *SecretJSON, configLoader *ConfigLoader, fileManager FileManagerInterface) CollectionsViewModel {
	// Total pages = collections + 1 (settings page)
	totalPages := len(collections) + 1

	m := CollectionsViewModel{
		collections:   collections,
		config:        config,
		secret:        secret,
		configLoader:  configLoader,
		fileManager:   fileManager,
		currentPage:   0,
		cursor:        0,
		viewportStart: 0,
		maxVisible:    10,
		quitting:      false,
		styles:        DefaultStyles(),
		searchMode:    false,
		searchQuery:   "",
		filteredList:  []RequestItem{},
		totalPages:    totalPages,
	}

	return m
}

func (m CollectionsViewModel) Init() tea.Cmd {
	return nil
}

func (m CollectionsViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.searchMode {
				m.searchMode = false
				m.searchQuery = ""
				m.filteredList = []RequestItem{}
				m.cursor = 0
				m.viewportStart = 0
				return m, nil
			}
			*m.selected = ExitSignal
			m.quitting = true
			return m, tea.Quit

		case "esc":
			if m.searchMode {
				m.searchMode = false
				m.searchQuery = ""
				m.filteredList = []RequestItem{}
				m.cursor = 0
				m.viewportStart = 0
				return m, nil
			}
			*m.selected = ExitSignal
			m.quitting = true
			return m, tea.Quit

		case "/":
			if !m.searchMode {
				m.searchMode = true
				m.searchQuery = ""
				m.filteredList = []RequestItem{}
				m.cursor = 0
				m.viewportStart = 0
				return m, nil
			}

		case "backspace":
			if m.searchMode && len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.updateFilteredList()
				m.cursor = 0
				m.viewportStart = 0
				return m, nil
			}

		case "left", "h":
			if m.searchMode {
				return m, nil
			}
			if m.currentPage > 0 {
				m.currentPage--
			} else {
				m.currentPage = m.totalPages - 1
			}
			m.cursor = 0
			m.viewportStart = 0

		case "right", "l":
			if m.searchMode {
				return m, nil
			}
			if m.currentPage < m.totalPages-1 {
				m.currentPage++
			} else {
				m.currentPage = 0
			}
			m.cursor = 0
			m.viewportStart = 0

		case "up", "k":
			items := m.getActiveList()
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.viewportStart+2 && m.viewportStart > 0 {
					m.viewportStart--
				}
			} else {
				m.cursor = len(items) - 1
				if len(items) > m.maxVisible {
					m.viewportStart = len(items) - m.maxVisible
				} else {
					m.viewportStart = 0
				}
			}

		case "down", "j":
			items := m.getActiveList()
			if m.cursor < len(items)-1 {
				m.cursor++
				if m.cursor >= m.viewportStart+m.maxVisible-2 {
					m.viewportStart++
				}
			} else {
				m.cursor = 0
				m.viewportStart = 0
			}

		case "enter":
			items := m.getActiveList()
			if len(items) > 0 && m.cursor < len(items) {
				// Check if we're on settings page
				if m.isSettingsPage() {
					settingsItem := m.getSettingsItems()[m.cursor]
					result := fmt.Sprintf("settings|%s", settingsItem.Key)
					*m.selected = result
					m.quitting = true
					return m, tea.Quit
				}

				// Regular request selection
				selectedItem := items[m.cursor]
				result := fmt.Sprintf("%s|%s", m.collections[m.currentPage].Name, selectedItem.Name)
				*m.selected = result
				m.quitting = true
				return m, tea.Quit
			}

		default:
			if m.searchMode {
				key := msg.String()
				if len(key) == 1 {
					m.searchQuery += key
					m.updateFilteredList()
					m.cursor = 0
					m.viewportStart = 0
					return m, nil
				}
			}
		}
	}

	return m, nil
}

func (m CollectionsViewModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Header with tabs
	var tabViews []string
	for i, collection := range m.collections {
		if i == m.currentPage {
			tabViews = append(tabViews, m.styles.Text(fmt.Sprintf("[ %s ]", collection.Name), m.styles.SelectedTitleColor))
		} else {
			tabViews = append(tabViews, m.styles.Text(fmt.Sprintf("  %s  ", collection.Name), m.styles.MutedTitleColor))
		}
	}

	// Add settings tab
	settingsPageIndex := len(m.collections)
	if settingsPageIndex == m.currentPage {
		tabViews = append(tabViews, m.styles.Text("[ settings ⚙️ ]", m.styles.SettingsSelectedTitleColor))
	} else {
		tabViews = append(tabViews, m.styles.Text("  settings ⚙️  ", m.styles.SettingsTitleColor))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tabViews...))
	b.WriteString("\n\n")

	// Show search box if in search mode
	if m.searchMode {
		searchBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(m.styles.SearchBoxColor).
			Padding(0, 1).
			Width(70).
			Foreground(m.styles.SearchTextColor)

		searchText := fmt.Sprintf("🔍 Search: %s", m.searchQuery)
		if m.searchQuery == "" {
			searchText = "🔍 Search: (type to search...)"
		}
		b.WriteString(searchBox.Render(searchText))
		b.WriteString("\n\n")
	}

	// Current page items
	items := m.getActiveList()
	if len(items) == 0 {
		if m.searchMode {
			b.WriteString(m.styles.FooterStyle.Render("  No matches found\n"))
		} else {
			b.WriteString(m.styles.FooterStyle.Render("  No requests in this collection\n"))
		}
	} else {
		visibleEnd := m.viewportStart + m.maxVisible
		if visibleEnd > len(items) {
			visibleEnd = len(items)
		}

		if m.viewportStart > 0 {
			b.WriteString(m.styles.FooterStyle.Render("  ⬆ More items above..."))
			b.WriteString("\n\n")
		}

		// Check if we're on settings page
		isSettings := m.isSettingsPage()

		for i := m.viewportStart; i < visibleEnd; i++ {
			var itemBox lipgloss.Style
			var borderColor lipgloss.Color

			if m.cursor == i {
				if isSettings {
					borderColor = m.styles.SettingsSelectedTitleColor
				} else {
					borderColor = m.styles.SelectedTitleColor
				}
				itemBox = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(borderColor).
					Padding(0, 1).
					Width(70).
					MarginLeft(2)
			} else {
				if isSettings {
					borderColor = m.styles.SettingsBorderColor
				} else {
					borderColor = m.styles.MutedBorderColor
				}
				itemBox = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(borderColor).
					Padding(0, 1).
					Width(70)
			}

			titleStyle := lipgloss.NewStyle().Bold(true)
			var valueColor lipgloss.Color
			if m.cursor == i {
				if isSettings {
					titleStyle = titleStyle.Foreground(m.styles.SettingsSelectedTitleColor)
					valueColor = m.styles.SettingsValueColor
				} else {
					titleStyle = titleStyle.Foreground(m.styles.SelectedTitleColor)
					valueColor = m.styles.FooterColor
				}
			} else {
				if isSettings {
					titleStyle = titleStyle.Foreground(m.styles.SettingsTitleColor)
					valueColor = m.styles.SettingsValueColor
				} else {
					titleStyle = titleStyle.Foreground(m.styles.MutedTitleColor)
					valueColor = m.styles.MutedTitleColor
				}
			}

			valueStyle := lipgloss.NewStyle().
				Foreground(valueColor).
				Width(66).
				Italic(true)

			var content string

			if isSettings {
				// Render settings item
				settingsItem := m.getSettingsItems()[i]
				var titleText string
				if m.searchMode && m.searchQuery != "" {
					titleText = m.highlightMatches(settingsItem.Label, m.searchQuery)
				} else {
					titleText = settingsItem.Label
				}

				content = fmt.Sprintf("%s\n%s",
					titleStyle.Render(titleText),
					valueStyle.Render(settingsItem.Value),
				)
			} else {
				// Render request item
				item := items[i]
				var titleText string
				if m.searchMode && m.searchQuery != "" {
					titleText = m.highlightMatches(item.Name, m.searchQuery)
				} else {
					titleText = item.Name
				}

				methodColor := m.getMethodColor(item.Request.Method)
				methodStyle := lipgloss.NewStyle().
					Foreground(methodColor).
					Bold(true)

				displayURL := m.configLoader.ReplaceVariables(item.Request.URL, m.config)

				content = fmt.Sprintf("%s\n%s %s",
					titleStyle.Render(titleText),
					methodStyle.Render(item.Request.Method),
					valueStyle.Render(displayURL),
				)
			}

			b.WriteString(itemBox.Render(content))
			b.WriteString("\n")
		}

		if visibleEnd < len(items) {
			b.WriteString("\n")
			b.WriteString(m.styles.FooterStyle.Render("  ⬇ More items below..."))
		}
	}

	// Footer
	b.WriteString("\n")
	var helpText string
	if m.searchMode {
		helpText = "  type to search • ↑↓/jk navigate • enter select • esc cancel"
	} else {
		helpText = "  / search • ↑↓/jk navigate • enter select • q/esc quit"
		if m.totalPages > 1 {
			helpText = "  / search • ←→/hl switch • ↑↓/jk navigate • enter select • q/esc quit"
		}
	}
	b.WriteString(m.styles.FooterStyle.Render(helpText + "\n"))

	return b.String()
}

func (m CollectionsViewModel) getActiveList() []RequestItem {
	if m.searchMode && m.searchQuery != "" {
		return m.filteredList
	}

	// Check if we're on settings page
	if m.isSettingsPage() {
		// Convert settings items to RequestItem format for display
		settingsItems := m.getSettingsItems()
		requestItems := make([]RequestItem, len(settingsItems))
		for i, item := range settingsItems {
			requestItems[i] = RequestItem{
				Name: item.Label,
				Request: &RequestJSON{
					URL: item.Value,
				},
			}
		}
		return requestItems
	}

	if m.currentPage < len(m.collections) {
		return m.collections[m.currentPage].Requests
	}
	return []RequestItem{}
}

func (m CollectionsViewModel) isSettingsPage() bool {
	return m.currentPage == len(m.collections)
}

type SettingsItem struct {
	Key   string
	Label string
	Value string
}

func (m CollectionsViewModel) getSettingsItems() []SettingsItem {
	items := []SettingsItem{
		{
			Key:   "baseUrl",
			Label: "Base URL",
			Value: m.config.BaseUrl,
		},
		{
			Key:   "jwt",
			Label: "JWT Token",
			Value: m.configLoader.GetJWT(m.secret),
		},
		{
			Key:   "timeout",
			Label: "Timeout (seconds)",
			Value: fmt.Sprintf("%d", m.config.GetTimeout()),
		},
	}
	return items
}

func (m *CollectionsViewModel) updateFilteredList() {
	if m.searchQuery == "" {
		m.filteredList = []RequestItem{}
		return
	}

	currentList := m.collections[m.currentPage].Requests
	m.filteredList = []RequestItem{}

	query := strings.ToLower(m.searchQuery)

	for _, item := range currentList {
		nameLower := strings.ToLower(item.Name)
		urlLower := strings.ToLower(item.Request.URL)

		if fuzzyMatch(nameLower, query) || fuzzyMatch(urlLower, query) {
			m.filteredList = append(m.filteredList, item)
		}
	}
}

func fuzzyMatch(text, query string) bool {
	if query == "" {
		return true
	}

	textIdx := 0
	queryIdx := 0

	for textIdx < len(text) && queryIdx < len(query) {
		if text[textIdx] == query[queryIdx] {
			queryIdx++
		}
		textIdx++
	}

	return queryIdx == len(query)
}

func (m CollectionsViewModel) highlightMatches(text, query string) string {
	if query == "" {
		return text
	}

	highlightStyle := lipgloss.NewStyle().
		Background(m.styles.HighlightBgColor).
		Foreground(m.styles.HighlightFgColor)

	textLower := strings.ToLower(text)
	queryLower := strings.ToLower(query)

	var result strings.Builder
	textIdx := 0
	queryIdx := 0

	for textIdx < len(text) {
		if queryIdx < len(queryLower) && textLower[textIdx] == queryLower[queryIdx] {
			result.WriteString(highlightStyle.Render(string(text[textIdx])))
			queryIdx++
		} else {
			result.WriteByte(text[textIdx])
		}
		textIdx++
	}

	return result.String()
}

func (m CollectionsViewModel) getMethodColor(method string) lipgloss.Color {
	switch method {
	case "GET":
		return m.styles.AquamarineColor
	case "POST":
		return m.styles.PeachColor
	case "PUT":
		return m.styles.ThistleColor
	case "DELETE":
		return m.styles.CoralColor
	case "PATCH":
		return m.styles.OrchidColor
	default:
		return m.styles.MutedTitleColor
	}
}

func CollectionsView(collections []Collection, config *ConfigJSON, secret *SecretJSON, configLoader *ConfigLoader, fileManager FileManagerInterface, selected *string) {
	m := NewCollectionsViewModel(collections, config, secret, configLoader, fileManager)
	m.selected = selected

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("CollectionsView -> ", err)
		os.Exit(1)
	}
}
