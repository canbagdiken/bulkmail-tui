package app

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	screen       int
	viewport     viewport.Model
	app          *App
	delayInput   textinput.Model
	width        int
	height       int
	confirmStart bool
	help         help.Model
	keys         keyMap
}

func (m *model) Init() tea.Cmd {
	ti := textinput.New()
	ti.Placeholder = "Delay seconds"
	ti.CharLimit = 10
	ti.Width = 20

	m.delayInput = ti
	m.width = 80
	m.height = 20
	m.confirmStart = false
	m.help = help.New()
	m.keys = keyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Logs: key.NewBinding(
			key.WithKeys("1", "l"),
			key.WithHelp("l", "logs"),
		),
		Stats: key.NewBinding(
			key.WithKeys("2", "s"),
			key.WithHelp("s", "stats"),
		),
		Preferences: key.NewBinding(
			key.WithKeys("3", "p"),
			key.WithHelp("p", "preferences"),
		),
		Import: key.NewBinding(
			key.WithKeys("4", "i"),
			key.WithHelp("i", "import"),
		),
		Pending: key.NewBinding(
			key.WithKeys("5", "e"),
			key.WithHelp("e", "pending"),
		),
		Boot: key.NewBinding(
			key.WithKeys("b", "B"),
			key.WithHelp("B", "boot"),
		),
		Stop: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "abort"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		Clear: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear"),
		),
		Help: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("H", "toggle help"),
		),
	}
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m *model) renderScreen() {
	m.viewport.SetContent(m.getContent())
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		action := m.app.HandleKeyPress(
			msg.String(),
			m.screen,
			m.confirmStart,
			m.app.viewData.IsRunning,
			m.delayInput.Focused(),
			m.app.viewData.SelectedFile,
			m.app.viewData.ImportFiles,
			m.delayInput.Value(),
		)

		if action.ShouldQuit {
			return m, tea.Quit
		}

		if action.ScreenChanged {
			m.setScreen(action.SetScreen)
		}

		if action.BlurInput {
			m.delayInput.Blur()
		}

		if action.FocusInput {
			m.delayInput.SetValue(fmt.Sprintf("%d", m.app.viewData.DelaySeconds))
			m.delayInput.Focus()
		}

		if action.ClearLogs {
			m.app.ClearLogs()
		}

		if action.ShowConfirm {
			m.confirmStart = true
		}

		if action.StopMail {
			m.app.stopCh <- true
			m.app.booted = false
		}

		if action.StartMail {
			m.app.booted = true
			m.confirmStart = false
		}

		if action.CancelConfirm {
			m.confirmStart = false
		}

		if action.FileSelected >= 0 {
			m.app.viewData.SelectedFile = action.FileSelected
			m.app.viewData.ImportContent = m.generateImportContent()
		}

		if action.UpdateDelay > 0 {
			m.app.delaySeconds = action.UpdateDelay
			m.app.viewData.DelaySeconds = action.UpdateDelay
		}

		if action.ImportFile != "" {
			err := m.app.ImportEmailsFromFile(action.ImportFile)
			if err != nil {
				m.app.addLog(fmt.Sprintf("Error importing %s: %v", action.ImportFile, err))
			} else {
				m.setScreen(0)
			}
		}

		if action.ToggleHelp {
			m.help.ShowAll = !m.help.ShowAll
		}

		if action.ScrollUp {
			m.viewport.LineUp(1)
		}

		if action.ScrollDown {
			m.viewport.LineDown(1)
		}

		if m.screen == 2 {
			m.delayInput, cmd = m.delayInput.Update(msg)
		}
		m.renderScreen()

	case tickMsg:
		m.renderScreen()
		if m.screen == 0 {
			m.viewport.GotoBottom()
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })

	case tea.WindowSizeMsg:
		m.width = msg.Width
		if m.width < 80 {
			m.width = 80
		}
		height := msg.Height - 5
		if height < 10 {
			height = 10
		}
		m.height = height
		m.viewport = viewport.New(m.width, height)
		m.renderScreen()
	}
	return m, cmd
}

func (m *model) setScreen(screen int) {
	m.screen = screen
	if screen == 3 {
		files, err := os.ReadDir(".")
		if err == nil {
			m.app.viewData.ImportFiles = []string{}
			for _, file := range files {
				if !file.IsDir() {
					m.app.viewData.ImportFiles = append(m.app.viewData.ImportFiles, file.Name())
				}
			}
		}
		m.app.viewData.SelectedFile = 0
		m.app.viewData.ImportContent = m.generateImportContent()
	}
}

func (m model) generateImportContent() string {
	content := "Select a file to import emails:\n\n"
	for i, file := range m.app.viewData.ImportFiles {
		if i == m.app.viewData.SelectedFile {
			content += "> " + file + "\n"
		} else {
			content += "  " + file + "\n"
		}
	}
	content += "\nUse Up/Down to select, Enter to import"
	return content
}

func (m model) getContent() string {
	var content string
	switch m.screen {
	case 0:
		boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Width(m.width - 4).Background(lipgloss.Color("235"))
		content = boxStyle.Render(m.app.viewData.LogsContent)
	case 1:
		content = m.app.viewData.StatsContent
	case 2:
		content += "Delay: " + m.delayInput.View() + "\n"
		content += "Enter to set"
	case 3:
		content = m.app.viewData.ImportContent
	case 4:
		content = m.app.viewData.PendingContent
	}
	if m.confirmStart {
		content += "\n\nConfirm start mail sending? Press y to start, n to cancel"
	}
	return content
}

func (m model) View() string {
	activeTabStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("220")).Bold(true).Padding(0, 1)
	inactiveTabStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Background(lipgloss.Color("236")).Padding(0, 1)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(m.width).BorderBottom(true).BorderStyle(lipgloss.NormalBorder()).Background(lipgloss.Color("0"))
	viewportStyle := lipgloss.NewStyle().Background(lipgloss.Color("0"))

	tabs := m.app.viewData.TabNames
	var tabLine string
	for i, tab := range tabs {
		if i == m.screen {
			tabLine += activeTabStyle.Render(tab) + " "
		} else {
			tabLine += inactiveTabStyle.Render(tab) + " "
		}
	}

	var statusText string
	var statusStyle lipgloss.Style
	if m.app.viewData.IsRunning {
		statusText = fmt.Sprintf(" %s ", m.app.viewData.StatusText)
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("196")).Bold(true)
	} else {
		statusText = fmt.Sprintf(" %s ", m.app.viewData.StatusText)
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("240"))
	}
	pendingText := fmt.Sprintf(" Pending: %d ", m.app.viewData.PendingCount)
	pendingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Background(lipgloss.Color("0"))

	status := statusStyle.Render(statusText) + pendingStyle.Render(pendingText)

	placeWidth := m.width - len(tabLine)
	if placeWidth < 0 {
		placeWidth = 0
	}
	fullTabLine := lipgloss.JoinHorizontal(lipgloss.Left, tabLine, lipgloss.PlaceHorizontal(placeWidth, lipgloss.Right, status))
	tabLine = borderStyle.Render(fullTabLine) + "\n"

	view := viewportStyle.Render(m.viewport.View())

	if m.help.ShowAll {
		helpModal := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Background(lipgloss.Color("0")).Foreground(lipgloss.Color("15")).Render(m.help.View(m.keys))
		helpModal = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpModal, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceForeground(lipgloss.Color("0")))
		view = lipgloss.JoinVertical(lipgloss.Left, view, helpModal)
	}

	if m.confirmStart {
		modalStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Background(lipgloss.Color("0")).Foreground(lipgloss.Color("15"))
		modalContent := modalStyle.Render("Confirm start mail sending?\n\nPress y to start, n to cancel")
		modal := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalContent, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceForeground(lipgloss.Color("0")))
		view = lipgloss.JoinVertical(lipgloss.Left, view, modal)
	}

	helpLine := ""
	if !m.help.ShowAll {
		helpLine = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("0")).Render(m.help.View(m.keys))
	}

	return lipgloss.NewStyle().Padding(1, 1).Render(tabLine + view + helpLine)
}

// RunTUI starts the Terminal User Interface
func RunTUI(application *App) error {
	m := model{app: application}
	p := tea.NewProgram(&m)
	_, err := p.Run()
	return err
}
