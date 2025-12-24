package models

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Width  int
	Height int
	Table  table.Model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			// switch Section
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Table.SetHeight(m.Height - 6)

		tableWidth := (m.Width / 2) - 4
		columns := []table.Column{
			{Title: "WeaponName", Width: int(float64(tableWidth) * 0.35)},
			{Title: "WeaponType", Width: int(float64(tableWidth) * 0.15)},
			{Title: "TargetType", Width: int(float64(tableWidth) * 0.10)},
			{Title: "MaxVelocity", Width: int(float64(tableWidth) * 0.10)},
			{Title: "Min-Max Range", Width: int(float64(tableWidth) * 0.20)},
		}
		m.Table.SetColumns(columns)
	}

	// Update table
	m.Table, cmd = m.Table.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	var content string
	// Divide home tab vertically
	leftSection := lipgloss.NewStyle().
		Width(m.Width / 2).
		Height(m.Height - 5).
		Render("Left Section Content")

	rightSection := lipgloss.NewStyle().
		Width(m.Width / 2).
		Height(m.Height - 5).
		Render(m.Table.View())

	content = lipgloss.JoinHorizontal(lipgloss.Top, leftSection, rightSection)
	// Help text
	help := lipgloss.NewStyle().
		Width(m.Width).
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("236")).
		Render("Tab: Switch Section | q: Quit")
	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		help,
	)
}

func InitialModel(missiles []WeaponData) Model {
	// Create table columns
	columns := []table.Column{
		// {Title: "WeaponFileName", Width: 20},
		{Title: "WeaponName", Width: 20},
		{Title: "WeaponType", Width: 20},
		{Title: "TargetType", Width: 20},
		{Title: "MaxVelocity", Width: 20},
		{Title: "Min-Max Range", Width: 20},
	}

	rows := []table.Row{}

	for _, missile := range missiles {
		if missile.WeaponProperties.MinLaunchRange != "" && missile.WeaponProperties.MaxLaunchRange != "" {
			rows = append(rows, []string{
				missile.WeaponName,
				missile.WeaponType,
				missile.TargetType,
				missile.WeaponProperties.MaxVelocity,
				"(" + missile.WeaponProperties.MinLaunchRange + " - " + missile.WeaponProperties.MaxLaunchRange + ")",
			})
		} else {
			rows = append(rows, []string{
				missile.WeaponName,
				missile.WeaponType,
				missile.TargetType,
				missile.WeaponProperties.MaxVelocity,
				"",
			})
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		// BorderStyle(lipgloss.NormalBorder()).
		// BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		// Foreground(lipgloss.Color("229")).
		// Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return Model{
		Table: t,
	}
}
