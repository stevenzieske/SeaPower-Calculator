package models

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"os"
	"strings"
)

type Model struct {
	Width             int
	Height            int
	Table             *table.Table
	AllTableRows      [][]string
	FilteredTableRows [][]string
	TableCursor       int
	SearchInput       textinput.Model
	Renderer          *lipgloss.Renderer
	TableScrollOffset int
	SelectedWeapons   []string
	Focused           string
	WeaponData        []WeaponData
}

type HeaderColumn struct {
	Title string
	Width float32
}

var headers = []HeaderColumn{
	{Title: "Name", Width: 0.30},
	{Title: "FileName", Width: 0.20},
	{Title: "Type", Width: 0.10},
	{Title: "Target", Width: 0.10},
	{Title: "Velocity", Width: 0.10},
	{Title: "Range", Width: 0.10},
}

func (m *Model) filterRows() {
	searchString := strings.ToLower(strings.TrimSpace(m.SearchInput.Value()))
	if searchString == "" {
		m.FilteredTableRows = m.AllTableRows
		return
	}

	m.FilteredTableRows = [][]string{}
	for _, row := range m.AllTableRows {
		if len(row) > 0 && strings.Contains(strings.ToLower(row[0]), searchString) {
			m.FilteredTableRows = append(m.FilteredTableRows, row)
		}
	}

}

func getHeaderTitles(headers []HeaderColumn) []string {
	titles := make([]string, len(headers))
	for i, h := range headers {
		titles[i] = strings.ToUpper(h.Title)
	}
	return titles
}

func (m *Model) rebuildTable() {
	headerTitles := getHeaderTitles(headers)
	tableWidth := m.Width/2 - 8

	sectionHeight := m.Height - 3 - 2 - 3 - 1 // Account for help, section borders, search, gap
	maxRows := sectionHeight - 4              // Subtract table borders and header

	if maxRows < 1 {
		maxRows = 1
	}

	visibleRows := make([][]string, maxRows)
	for i := 0; i < maxRows; i++ {
		rowIdx := m.TableScrollOffset + i
		if rowIdx < len(m.FilteredTableRows) {
			visibleRows[i] = m.FilteredTableRows[rowIdx]
		} else {
			visibleRows[i] = make([]string, len(headers))
		}
	}

	m.Table = table.New().
		Border(lipgloss.RoundedBorder()).
		Wrap(false).
		Headers(headerTitles...).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle()

			if col < len(headers) {
				style = style.Width(int(float32(tableWidth) * headers[col].Width))
			}

			actualRowIdx := m.TableScrollOffset + row
			if actualRowIdx == m.TableCursor && actualRowIdx < len(m.FilteredTableRows) {
				style = style.
					Background(lipgloss.Color("240")).
					Foreground(lipgloss.Color("230")).
					Bold(true)
			}

			return style
		}).
		Rows(visibleRows...)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	sectionHeight := m.Height - 3 - 2 - 3 - 1 // Help, section borders, search, gap
	maxRows := sectionHeight - 4              // Table borders and header
	if maxRows < 1 {
		maxRows = 1
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Focused == "search" {
			switch msg.String() {
			case "esc":
				m.Focused = "table"
				m.SearchInput.Blur()
				return m, nil
			}
			m.SearchInput, cmd = m.SearchInput.Update(msg)
			m.filterRows()
			m.TableCursor = 0
			m.TableScrollOffset = 0
			m.rebuildTable()
			return m, cmd
		}

		if m.Focused == "table" {
			switch msg.String() {
			case "up", "k":
				if m.TableCursor > 0 {
					m.TableCursor--
					if m.TableCursor < m.TableScrollOffset {
						m.TableScrollOffset = m.TableCursor
					}
				}
			case "down", "j":
				if m.TableCursor < len(m.FilteredTableRows)-1 {
					m.TableCursor++
					if m.TableCursor >= m.TableScrollOffset+maxRows {
						m.TableScrollOffset = m.TableCursor - maxRows + 1
					}
				}
			case "home", "g":
				m.TableCursor = 0
				m.TableScrollOffset = 0
			case "end", "G":
				m.TableCursor = len(m.FilteredTableRows) - 1
				m.TableScrollOffset = len(m.FilteredTableRows) - maxRows
				if m.TableScrollOffset < 0 {
					m.TableScrollOffset = 0
				}
			case "/":
				m.Focused = "search"
				m.SearchInput.Focus()
				m.TableCursor = 0
				m.TableScrollOffset = 0
				m.rebuildTable()
				return m, nil
			case "enter":
				m.SelectedWeapons = append(m.SelectedWeapons, m.FilteredTableRows[m.TableCursor][1])
			case "R":
				m.Focused = "table"
				m.SelectedWeapons = []string{}
				m.FilteredTableRows = m.AllTableRows
				m.SearchInput.SetValue("")
				m.TableCursor = 0
				m.TableScrollOffset = 0
				m.rebuildTable()
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	m.rebuildTable()

	return m, nil
}

func (m Model) View() string {
	helpText := "↑/↓ k/j: Navigate • g/G: Top/Bottom • R: Reset • q: Quit"
	helpText += fmt.Sprintf(" • Row %d/%d", m.TableCursor+1, len(m.FilteredTableRows))

	helpBar := lipgloss.NewStyle().
		Width(m.Width - 2).
		Foreground(lipgloss.Color("245")).
		Background(lipgloss.Color("235")).
		PaddingLeft(1).
		Render(helpText)

	leftSection := lipgloss.NewStyle().
		Width(m.Width/2 - 2).
		Height(m.Height - 2 - lipgloss.Height(helpBar)).
		Border(lipgloss.RoundedBorder()).
		AlignVertical(lipgloss.Top).
		Render(m.SelectedWeapons...)

	searchStyle := lipgloss.NewStyle().
		Width(m.Width/2-6).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(0, 1)

	searchBox := searchStyle.Render("Search: " + m.SearchInput.View())

	rightContent := lipgloss.JoinVertical(lipgloss.Left,
		searchBox,
		"",
		m.Table.Render(),
	)

	rightSection := lipgloss.NewStyle().
		Width(m.Width/2 - 2).
		Height(m.Height - 2 - lipgloss.Height(helpBar)).
		Border(lipgloss.RoundedBorder()).
		AlignVertical(lipgloss.Top).
		Render(rightContent)

	sections := lipgloss.JoinHorizontal(lipgloss.Top, leftSection, rightSection)

	content := lipgloss.JoinVertical(lipgloss.Left, sections, helpBar)

	return content
}

func InitialModel(weapons []WeaponData) Model {
	rows := [][]string{}
	for _, weapon := range weapons {
		rangeStr := ""
		if weapon.WeaponProperties.MinLaunchRange != "" && weapon.WeaponProperties.MaxLaunchRange != "" {
			rangeStr = weapon.WeaponProperties.MinLaunchRange + "-" + weapon.WeaponProperties.MaxLaunchRange
		}
		rows = append(rows, []string{
			weapon.WeaponName,
			weapon.WeaponFileName,
			weapon.WeaponType,
			weapon.TargetType,
			weapon.WeaponProperties.MaxVelocity,
			rangeStr,
		})
	}

	re := lipgloss.NewRenderer(os.Stdout)

	ti := textinput.New()

	m := Model{
		Renderer:          re,
		AllTableRows:      rows,
		FilteredTableRows: rows,
		TableCursor:       0,
		SearchInput:       ti,
		Focused:           "table",
		WeaponData:        weapons,
	}

	headerTitles := getHeaderTitles(headers)
	tableWidth := m.Width/2 - 8

	m.Table = table.New().
		Border(lipgloss.RoundedBorder()).
		Wrap(false).
		Headers(headerTitles...).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle()
			if col < len(headers) {
				style = style.Width(int(float32(tableWidth) * headers[col].Width))
			}

			if row == m.TableCursor && row < len(rows) {
				style = style.
					Background(lipgloss.Color("240")).
					Foreground(lipgloss.Color("230")).
					Bold(true)
			}

			return style
		}).
		Rows(rows...)

	return m
}
