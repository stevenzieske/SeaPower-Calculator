package tui

import (
	"log"

	"github.com/stevenzieske/SeaPower-Calculator/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

func RenderTUI(weapons []models.WeaponData) {
	p := tea.NewProgram(
		models.InitialModel(weapons),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Printf("Error: %v", err)
	}
}
