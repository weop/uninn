package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"uninn/internal/ui"
)

func main() {
	// Check if running with proper permissions hint
	if os.Geteuid() == 0 {
		fmt.Println("Warning: Running as root is not recommended. The app will request permissions when needed.")
	}

	model := ui.NewModel()
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}