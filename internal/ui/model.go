package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"uninn/internal/apt"
	"uninn/internal/appimage"
	"uninn/internal/flatpak"
	"uninn/internal/pacman"
	"uninn/internal/rpm"
	"uninn/internal/snap"
	"uninn/pkg"
)

type state int

const (
	stateList state = iota
	stateConfirm
	stateUninstalling
	stateDone
)

type packageItem struct {
	pkg.Package
}

func (i packageItem) FilterValue() string { return i.Name }
func (i packageItem) Title() string       { return i.Name }
func (i packageItem) Description() string {
	return fmt.Sprintf("[%s] %s - %s", i.Manager, i.Version, i.Size)
}

type Model struct {
	list          list.Model
	allPackages   []pkg.Package
	detectors     map[pkg.PackageManager]pkg.PackageDetector
	state         state
	selectedPkg   *pkg.Package
	searchInput   textinput.Model
	err           error
	width         int
	height        int
	confirmDialog string
	loading       bool
	loadingMsg    string
}

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Padding(0, 1)

	confirmStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1, 2).
		MarginTop(1)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")).
		Bold(true)
)

func NewModel() *Model {
	searchInput := textinput.New()
	searchInput.Placeholder = "Search packages..."
	searchInput.Focus()

	detectors := map[pkg.PackageManager]pkg.PackageDetector{
		pkg.APT:      apt.New(),
		pkg.Pacman:   pacman.New(),
		pkg.Flatpak:  flatpak.New(),
		pkg.Snap:     snap.New(),
		pkg.AppImage: appimage.New(),
		pkg.RPM:      rpm.New(),
	}

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Universal Package Uninstaller"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return &Model{
		list:        l,
		detectors:   detectors,
		searchInput: searchInput,
		state:       stateList,
		loading:     true,
		loadingMsg:  "Loading packages...",
	}
}

func (m *Model) Init() tea.Cmd {
	return m.loadPackages
}

func (m *Model) loadPackages() tea.Msg {
	var allPackages []pkg.Package
	
	for _, detector := range m.detectors {
		if detector.IsAvailable() {
			packages, err := detector.ListPackages()
			if err != nil {
				// Continue loading other package managers
				continue
			}
			allPackages = append(allPackages, packages...)
		}
	}
	
	return packagesLoadedMsg{packages: allPackages}
}

type packagesLoadedMsg struct {
	packages []pkg.Package
}

type uninstallCompleteMsg struct {
	success bool
	err     error
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
		return m, nil

	case packagesLoadedMsg:
		m.loading = false
		m.allPackages = msg.packages
		items := make([]list.Item, len(m.allPackages))
		for i, p := range m.allPackages {
			items[i] = packageItem{p}
		}
		m.list.SetItems(items)
		return m, nil

	case uninstallCompleteMsg:
		if msg.success {
			m.confirmDialog = successStyle.Render("✓ Package uninstalled successfully!")
			m.state = stateDone
			// Reload packages
			return m, tea.Sequence(
				tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
					return tea.KeyMsg{Type: tea.KeyEsc}
				}),
				m.loadPackages,
			)
		} else {
			m.err = msg.err
			m.confirmDialog = errorStyle.Render(fmt.Sprintf("✗ Failed to uninstall: %v", msg.err))
			m.state = stateDone
			return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEsc}
			})
		}

	case tea.KeyMsg:
		switch m.state {
		case stateList:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				if selectedItem, ok := m.list.SelectedItem().(packageItem); ok {
					m.selectedPkg = &selectedItem.Package
					m.state = stateConfirm
					m.confirmDialog = fmt.Sprintf("Are you sure you want to uninstall %s?\n\nPress 'y' to confirm or 'n' to cancel", selectedItem.Name)
				}
				return m, nil
			}

		case stateConfirm:
			switch msg.String() {
			case "y", "Y":
				m.state = stateUninstalling
				m.confirmDialog = "Uninstalling..."
				return m, m.uninstallPackage
			case "n", "N", "esc":
				m.state = stateList
				m.confirmDialog = ""
				return m, nil
			}

		case stateDone:
			m.state = stateList
			m.confirmDialog = ""
			return m, nil
		}
	}

	if m.state == stateList && !m.loading {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) uninstallPackage() tea.Msg {
	if m.selectedPkg == nil {
		return uninstallCompleteMsg{success: false, err: fmt.Errorf("no package selected")}
	}

	detector := m.detectors[m.selectedPkg.Manager]
	
	var err error
	if m.selectedPkg.Manager == pkg.AppImage {
		// For AppImages, directly remove the file
		err = os.Remove(m.selectedPkg.Path)
	} else {
		err = detector.Uninstall(m.selectedPkg.Name)
	}

	return uninstallCompleteMsg{
		success: err == nil,
		err:     err,
	}
}

func (m *Model) View() string {
	if m.loading {
		return fmt.Sprintf("\n\n   %s\n\n", m.loadingMsg)
	}

	switch m.state {
	case stateList:
		return fmt.Sprintf("%s\n%s", m.list.View(), m.helpView())
	case stateConfirm, stateUninstalling, stateDone:
		return fmt.Sprintf("%s\n\n%s\n\n%s", 
			m.list.View(), 
			confirmStyle.Render(m.confirmDialog),
			m.helpView())
	default:
		return m.list.View()
	}
}

func (m *Model) helpView() string {
	if m.state == stateList {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			"↑/↓: navigate • enter: uninstall • /: search • q: quit",
		)
	}
	return ""
}