package tui

import (
	"log"
	"seapower_calculator/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

func RenderTUI(missiles []models.WeaponData) {
	p := tea.NewProgram(
		models.InitialModel(missiles),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Printf("Error: %v", err)
	}
}
