package models

import (
	"fmt"
	"os"
	"seapower_calculator/internal/helper"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type Model struct {
	Width                 int
	Height                int
	Table                 *table.Table
	AllTableRows          [][]string
	FilteredTableRows     [][]string
	TableCursor           int
	SearchInput           textinput.Model
	RangeInput            textinput.Model
	RangeInputs           []textinput.Model // Range inputs for each weapon in delta time mode
	Renderer              *lipgloss.Renderer
	TableScrollOffset     int
	SelectedWeapons       []WeaponData
	SelectionScrollOffset int
	SelectionCursor       int
	Focused               string
	WeaponData            []WeaponData
	CalculationMode       string // "flightTime" or "deltaTime"
	IsSorted              bool   // Track if selections are sorted
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

// Calculate how many weapon items fit in the left section based on mode and height
func (m *Model) getMaxVisibleItems() int {
	// Section height calculation (same as in View())
	// m.Height - 2 (top/bottom margin) - 1 (helpBar) = m.Height - 3
	sectionHeight := m.Height - 3

	// Account for section borders (top and bottom)
	contentHeight := sectionHeight - 2

	// Account for header box with border (~4 lines)
	contentHeight -= 4

	// Account for scroll indicators (1 line each when visible, but reserve space)
	contentHeight -= 2

	var itemHeight int
	if m.CalculationMode == "flightTime" {
		itemHeight = 10 // Each weapon item takes ~10 lines
	} else if m.CalculationMode == "deltaTime" {
		itemHeight = 14 // Each weapon item takes ~14 lines in delta time mode
	}

	maxVisibleItems := contentHeight / itemHeight

	if maxVisibleItems < 1 {
		maxVisibleItems = 1
	}

	return maxVisibleItems
}

func (m *Model) renderSelections() string {
	if len(m.SelectedWeapons) == 0 {
		return "No weapons selected"
	}

	// Calculate how many items fit based on current mode and window height
	maxVisibleItems := m.getMaxVisibleItems()

	// Calculate visible range
	startIdx := m.SelectionScrollOffset
	endIdx := startIdx + maxVisibleItems
	if endIdx > len(m.SelectedWeapons) {
		endIdx = len(m.SelectedWeapons)
	}

	selections := []string{}

	// Show indicator if there are items above
	if startIdx > 0 {
		indicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true).
			Render(fmt.Sprintf("▲ %d more above", startIdx))
		selections = append(selections, indicator)
	}

	// Mode indicator and calculations
	var headerText string
	var flightTimeHours float64 // Store for calculating ranges of other weapons in flightTime mode

	if m.CalculationMode == "flightTime" {
		// Calculate flight time if range is entered and valid
		flightTimeText := "Flight time: N/A"
		enteredRange := m.RangeInput.Value()
		if enteredRange != "" && len(m.SelectedWeapons) > 0 {
			// Parse entered range
			rangeValue, errRange := strconv.ParseFloat(enteredRange, 64)

			// Parse velocity from first weapon
			velocityStr := m.SelectedWeapons[0].WeaponProperties.MaxVelocity
			velocityValue, errVel := strconv.ParseFloat(velocityStr, 64)

			// Parse min/max range
			minRangeStr := m.SelectedWeapons[0].WeaponProperties.MinLaunchRange
			maxRangeStr := m.SelectedWeapons[0].WeaponProperties.MaxLaunchRange
			minRange, errMin := strconv.ParseFloat(minRangeStr, 64)
			maxRange, errMax := strconv.ParseFloat(maxRangeStr, 64)

			// Only calculate if range is valid and within bounds
			if errRange == nil && errVel == nil && errMin == nil && errMax == nil {
				if rangeValue >= minRange && rangeValue <= maxRange {
					// Calculate flight time in hours
					flightTimeHours = helper.CalculateFlightTime(int64(velocityValue), int64(rangeValue))

					// Convert to hours, minutes, and seconds for display
					hours := int(flightTimeHours)
					remainingAfterHours := flightTimeHours - float64(hours)
					totalMinutes := remainingAfterHours * 60
					minutes := int(totalMinutes)
					seconds := int((totalMinutes - float64(minutes)) * 60)

					if hours > 0 {
						flightTimeText = fmt.Sprintf("Flight time: %dh %dm %ds", hours, minutes, seconds)
					} else if minutes > 0 {
						flightTimeText = fmt.Sprintf("Flight time: %dm %ds", minutes, seconds)
					} else {
						flightTimeText = fmt.Sprintf("Flight time: %ds", seconds)
					}
				}
			}
		}
		headerText = flightTimeText
	} else {
		headerText = "Mode: Delta Time Calculation"
	}

	header := lipgloss.NewStyle().
		Width(m.Width/2-4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Foreground(lipgloss.Color("226")).
		Bold(true).
		Padding(0, 1).
		Render(headerText)
	selections = append(selections, header)

	// In deltaTime mode, calculate all flight times and find maximum
	var flightTimes []float64
	var maxFlightTime float64
	if m.CalculationMode == "deltaTime" && len(m.RangeInputs) == len(m.SelectedWeapons) {
		flightTimes = make([]float64, len(m.SelectedWeapons))
		for i := range m.SelectedWeapons {
			rangeValue, errRange := strconv.ParseFloat(m.RangeInputs[i].Value(), 64)
			velocityValue, errVel := strconv.ParseFloat(m.SelectedWeapons[i].WeaponProperties.MaxVelocity, 64)

			if errRange == nil && errVel == nil && rangeValue > 0 && velocityValue > 0 {
				flightTimes[i] = helper.CalculateFlightTime(int64(velocityValue), int64(rangeValue))
				if flightTimes[i] > maxFlightTime {
					maxFlightTime = flightTimes[i]
				}
			}
		}
	}

	for i := startIdx; i < endIdx; i++ {
		selectedWeapon := m.SelectedWeapons[i]

		// Build base info
		info := fmt.Sprintf("NAME: %s\nFILENAME: %s\nTYPE: %s\nTARGET: %s\nVELOCITY: %s\nRANGE: %s",
			selectedWeapon.WeaponName,
			selectedWeapon.WeaponFileName,
			selectedWeapon.WeaponType,
			selectedWeapon.TargetType,
			selectedWeapon.WeaponProperties.MaxVelocity,
			selectedWeapon.WeaponProperties.MinLaunchRange+" - "+selectedWeapon.WeaponProperties.MaxLaunchRange,
		)

		// Mode-specific calculations
		if m.CalculationMode == "flightTime" {
			// For weapons after the first, calculate range using flight time from first weapon
			if i > 0 && flightTimeHours > 0 {
				// Parse velocity of this weapon
				velocityValue, err := strconv.ParseFloat(selectedWeapon.WeaponProperties.MaxVelocity, 64)
				if err == nil {
					// Calculate range for this weapon using the same flight time
					calculatedRange := helper.CalculateRange(int64(velocityValue), flightTimeHours)

					// Parse min/max range for validation
					minRange, errMin := strconv.ParseFloat(selectedWeapon.WeaponProperties.MinLaunchRange, 64)
					maxRange, errMax := strconv.ParseFloat(selectedWeapon.WeaponProperties.MaxLaunchRange, 64)

					// Display calculated range with validation
					rangeText := fmt.Sprintf("%.1f nm", calculatedRange)
					rangeColor := lipgloss.Color("255") // Default white

					if errMin == nil && errMax == nil {
						if calculatedRange < minRange || calculatedRange > maxRange {
							rangeText += " ⚠ OUT OF RANGE"
							rangeColor = lipgloss.Color("196") // Red
						} else {
							rangeText += " ✓"
							rangeColor = lipgloss.Color("46") // Green
						}
					}

					styledRangeText := lipgloss.NewStyle().
						Foreground(rangeColor).
						Bold(true).
						Render(rangeText)

					info += fmt.Sprintf("\nCALCULATED RANGE: %s", styledRangeText)
				}
			}
		} else if m.CalculationMode == "deltaTime" && maxFlightTime > 0 {
			var deltaTexts []string
			var deltaColor lipgloss.Color

			if m.IsSorted && i > 0 {
				// When sorted, show delta time to previous weapon
				deltaTime := flightTimes[i-1] - flightTimes[i]

				// Convert delta time to hours, minutes, seconds
				hours := int(deltaTime)
				remainingAfterHours := deltaTime - float64(hours)
				totalMinutes := remainingAfterHours * 60
				minutes := int(totalMinutes)
				seconds := int((totalMinutes - float64(minutes)) * 60)

				var deltaTextPrev string
				if hours > 0 {
					deltaTextPrev = fmt.Sprintf("Deploy after previous: %dh %dm %ds", hours, minutes, seconds)
				} else if minutes > 0 {
					deltaTextPrev = fmt.Sprintf("Deploy after previous: %dm %ds", minutes, seconds)
				} else {
					deltaTextPrev = fmt.Sprintf("Deploy after previous: %ds", seconds)
				}
				deltaTexts = append(deltaTexts, deltaTextPrev)

				// Also show time relative to first weapon
				deltaTimeFirst := flightTimes[0] - flightTimes[i]
				hours = int(deltaTimeFirst)
				remainingAfterHours = deltaTimeFirst - float64(hours)
				totalMinutes = remainingAfterHours * 60
				minutes = int(totalMinutes)
				seconds = int((totalMinutes - float64(minutes)) * 60)

				var deltaTextFirst string
				if hours > 0 {
					deltaTextFirst = fmt.Sprintf("Deploy after first: %dh %dm %ds", hours, minutes, seconds)
				} else if minutes > 0 {
					deltaTextFirst = fmt.Sprintf("Deploy after first: %dm %ds", minutes, seconds)
				} else {
					deltaTextFirst = fmt.Sprintf("Deploy after first: %ds", seconds)
				}
				deltaTexts = append(deltaTexts, deltaTextFirst)
				deltaColor = lipgloss.Color("46") // Green
			} else {
				// When not sorted, show delta time to slowest weapon
				deltaTime := maxFlightTime - flightTimes[i]

				// Convert delta time to hours, minutes, seconds
				hours := int(deltaTime)
				remainingAfterHours := deltaTime - float64(hours)
				totalMinutes := remainingAfterHours * 60
				minutes := int(totalMinutes)
				seconds := int((totalMinutes - float64(minutes)) * 60)

				if deltaTime == 0 {
					deltaTexts = append(deltaTexts, "Deploy: NOW (slowest weapon)")
					deltaColor = lipgloss.Color("226") // Yellow for the first to deploy
				} else {
					var deltaText string
					if hours > 0 {
						deltaText = fmt.Sprintf("Deploy after slowest: %dh %dm %ds", hours, minutes, seconds)
					} else if minutes > 0 {
						deltaText = fmt.Sprintf("Deploy after slowest: %dm %ds", minutes, seconds)
					} else {
						deltaText = fmt.Sprintf("Deploy after slowest: %ds", seconds)
					}
					deltaTexts = append(deltaTexts, deltaText)
					deltaColor = lipgloss.Color("46") // Green
				}
			}

			// Render all delta texts
			for _, deltaText := range deltaTexts {
				styledDeltaText := lipgloss.NewStyle().
					Foreground(deltaColor).
					Bold(true).
					Render(deltaText)

				info += fmt.Sprintf("\n%s", styledDeltaText)
			}
		}

		style := lipgloss.NewStyle().
			Padding(0, 1).
			Width(m.Width/2 - 4).
			Border(lipgloss.RoundedBorder()).
			MarginBottom(1)

		// Show range input based on mode
		if m.CalculationMode == "flightTime" && i == 0 {
			// Only first weapon gets range input in flightTime mode
			labelColor := lipgloss.Color("238") // Default gray
			enteredRange := m.RangeInput.Value()
			if enteredRange != "" {
				// Parse the entered range and weapon's min/max
				var enteredValue float64
				_, err := fmt.Sscanf(enteredRange, "%f", &enteredValue)

				var minRange, maxRange float64
				_, errMin := fmt.Sscanf(selectedWeapon.WeaponProperties.MinLaunchRange, "%f", &minRange)
				_, errMax := fmt.Sscanf(selectedWeapon.WeaponProperties.MaxLaunchRange, "%f", &maxRange)

				// Check if parsing failed (invalid input)
				if err != nil {
					labelColor = lipgloss.Color("196") // Red - invalid input
				} else if errMin == nil && errMax == nil {
					// If parsing succeeded, check if in range
					if enteredValue < minRange || enteredValue > maxRange {
						labelColor = lipgloss.Color("196") // Red - out of range
					} else {
						labelColor = lipgloss.Color("46") // Green - valid and in range
					}
				}
			}

			rangeInputStyle := lipgloss.NewStyle().
				Width(m.Width/2-12).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(labelColor).
				Padding(0, 1)
			rangeBox := rangeInputStyle.Render("Weapon Deploy Range (nm): " + m.RangeInput.View())
			info += "\n" + rangeBox

			style = style.
				Border(lipgloss.RoundedBorder()).
				Foreground(lipgloss.Color("230")).
				Bold(true)
		} else if m.CalculationMode == "deltaTime" && i < len(m.RangeInputs) {
			// All weapons get range input in deltaTime mode
			labelColor := lipgloss.Color("238") // Default gray
			enteredRange := m.RangeInputs[i].Value()
			if enteredRange != "" {
				// Parse the entered range and weapon's min/max
				var enteredValue float64
				_, err := fmt.Sscanf(enteredRange, "%f", &enteredValue)

				var minRange, maxRange float64
				_, errMin := fmt.Sscanf(selectedWeapon.WeaponProperties.MinLaunchRange, "%f", &minRange)
				_, errMax := fmt.Sscanf(selectedWeapon.WeaponProperties.MaxLaunchRange, "%f", &maxRange)

				// Check if parsing failed (invalid input)
				if err != nil {
					labelColor = lipgloss.Color("196") // Red - invalid input
				} else if errMin == nil && errMax == nil {
					// If parsing succeeded, check if in range
					if enteredValue < minRange || enteredValue > maxRange {
						labelColor = lipgloss.Color("196") // Red - out of range
					} else {
						labelColor = lipgloss.Color("46") // Green - valid and in range
					}
				}
			}

			rangeInputStyle := lipgloss.NewStyle().
				Width(m.Width/2-12).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(labelColor).
				Padding(0, 1)
			rangeBox := rangeInputStyle.Render("Deploy Range (nm): " + m.RangeInputs[i].View())
			info += "\n" + rangeBox
		}

		// Highlight the cursor selection with different color
		if i == m.SelectionCursor {
			style = style.
				BorderForeground(lipgloss.Color("62")) // Blue background for cursor
		}

		selections = append(selections, style.Render(info))
	}

	// Show indicator if there are items below
	if endIdx < len(m.SelectedWeapons) {
		remaining := len(m.SelectedWeapons) - endIdx
		indicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true).
			Render(fmt.Sprintf("▼ %d more below", remaining))
		selections = append(selections, indicator)
	}

	return lipgloss.JoinVertical(lipgloss.Left, selections...)
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

		if m.Focused == "leftSection" {
			maxVisibleItems := m.getMaxVisibleItems()

			switch msg.String() {
			case "up", "k":
				// Move cursor up
				if m.SelectionCursor > 0 {
					m.SelectionCursor--
					// Auto-scroll if cursor moves above visible area
					if m.SelectionCursor < m.SelectionScrollOffset {
						m.SelectionScrollOffset = m.SelectionCursor
					}
				}
			case "down", "j":
				// Move cursor down
				if m.SelectionCursor < len(m.SelectedWeapons)-1 {
					m.SelectionCursor++
					// Auto-scroll if cursor moves below visible area
					if m.SelectionCursor >= m.SelectionScrollOffset+maxVisibleItems {
						m.SelectionScrollOffset = m.SelectionCursor - maxVisibleItems + 1
					}
				}
			case "home", "g":
				m.SelectionCursor = 0
				m.SelectionScrollOffset = 0
			case "end", "G":
				m.SelectionCursor = len(m.SelectedWeapons) - 1
				m.SelectionScrollOffset = len(m.SelectedWeapons) - maxVisibleItems
				if m.SelectionScrollOffset < 0 {
					m.SelectionScrollOffset = 0
				}
			case "backspace", "delete":
				if len(m.SelectedWeapons) > 0 && m.SelectionCursor < len(m.SelectedWeapons) {
					// Remove the selected weapon
					m.SelectedWeapons = append(m.SelectedWeapons[:m.SelectionCursor], m.SelectedWeapons[m.SelectionCursor+1:]...)

					// In deltaTime mode, also remove the corresponding range input
					if m.CalculationMode == "deltaTime" && m.SelectionCursor < len(m.RangeInputs) {
						m.RangeInputs = append(m.RangeInputs[:m.SelectionCursor], m.RangeInputs[m.SelectionCursor+1:]...)
					}

					// Adjust cursor if needed
					if m.SelectionCursor >= len(m.SelectedWeapons) && m.SelectionCursor > 0 {
						m.SelectionCursor--
					}

					// Adjust scroll offset if needed
					if m.SelectionScrollOffset > 0 && m.SelectionCursor < m.SelectionScrollOffset {
						m.SelectionScrollOffset--
					}

					// Update range input if first item was deleted or list is empty
					if len(m.SelectedWeapons) == 0 {
						m.RangeInput.SetValue("")
					}

					// Reset sort state when removing weapons
					m.IsSorted = false
				}
			case "/":
				// Focus the range input field
				m.Focused = "rangeInput"
				if m.CalculationMode == "flightTime" {
					m.RangeInput.Focus()
					// Scroll to top to show the range input (on first item)
					m.SelectionCursor = 0
					m.SelectionScrollOffset = 0
				} else if m.CalculationMode == "deltaTime" && m.SelectionCursor < len(m.RangeInputs) {
					m.RangeInputs[m.SelectionCursor].Focus()
				}
				return m, nil
			case "s":
				// Sort selections
				if len(m.SelectedWeapons) > 1 {
					if m.CalculationMode == "flightTime" {
						// Sort by calculated range in flightTime mode (excluding first weapon)
						enteredRange := m.RangeInput.Value()
						if enteredRange != "" && len(m.SelectedWeapons) > 1 {
							rangeValue, errRange := strconv.ParseFloat(enteredRange, 64)
							velocityStr := m.SelectedWeapons[0].WeaponProperties.MaxVelocity
							velocityValue, errVel := strconv.ParseFloat(velocityStr, 64)
							minRangeStr := m.SelectedWeapons[0].WeaponProperties.MinLaunchRange
							maxRangeStr := m.SelectedWeapons[0].WeaponProperties.MaxLaunchRange
							minRange, errMin := strconv.ParseFloat(minRangeStr, 64)
							maxRange, errMax := strconv.ParseFloat(maxRangeStr, 64)

							if errRange == nil && errVel == nil && errMin == nil && errMax == nil {
								if rangeValue >= minRange && rangeValue <= maxRange {
									flightTimeHours := helper.CalculateFlightTime(int64(velocityValue), int64(rangeValue))

									// Create slice of weapons with their calculated ranges (excluding first)
									type weaponWithRange struct {
										weapon WeaponData
										range_ float64
									}
									weaponsWithRanges := make([]weaponWithRange, len(m.SelectedWeapons)-1)

									for i := 1; i < len(m.SelectedWeapons); i++ {
										vel, err := strconv.ParseFloat(m.SelectedWeapons[i].WeaponProperties.MaxVelocity, 64)
										if err == nil {
											calcRange := helper.CalculateRange(int64(vel), flightTimeHours)
											weaponsWithRanges[i-1] = weaponWithRange{m.SelectedWeapons[i], calcRange}
										} else {
											weaponsWithRanges[i-1] = weaponWithRange{m.SelectedWeapons[i], 0}
										}
									}

									// Sort by range (ascending)
									sort.Slice(weaponsWithRanges, func(i, j int) bool {
										return weaponsWithRanges[i].range_ < weaponsWithRanges[j].range_
									})

									// Update selected weapons with sorted order (keep first weapon first)
									for i := range weaponsWithRanges {
										m.SelectedWeapons[i+1] = weaponsWithRanges[i].weapon
									}
									m.IsSorted = true
									m.SelectionCursor = 0
									m.SelectionScrollOffset = 0
								}
							}
						}
					} else if m.CalculationMode == "deltaTime" && len(m.RangeInputs) == len(m.SelectedWeapons) {
						// Sort by delta time (flight time) in deltaTime mode
						type weaponWithTime struct {
							weapon     WeaponData
							rangeInput textinput.Model
							flightTime float64
						}
						weaponsWithTimes := make([]weaponWithTime, len(m.SelectedWeapons))

						for i := range m.SelectedWeapons {
							rangeValue, errRange := strconv.ParseFloat(m.RangeInputs[i].Value(), 64)
							velocityValue, errVel := strconv.ParseFloat(m.SelectedWeapons[i].WeaponProperties.MaxVelocity, 64)

							var flightTime float64
							if errRange == nil && errVel == nil && rangeValue > 0 && velocityValue > 0 {
								flightTime = helper.CalculateFlightTime(int64(velocityValue), int64(rangeValue))
							}

							weaponsWithTimes[i] = weaponWithTime{
								weapon:     m.SelectedWeapons[i],
								rangeInput: m.RangeInputs[i],
								flightTime: flightTime,
							}
						}

						// Sort by flight time (descending - longest first)
						sort.Slice(weaponsWithTimes, func(i, j int) bool {
							return weaponsWithTimes[i].flightTime > weaponsWithTimes[j].flightTime
						})

						// Update selected weapons and range inputs with sorted order
						for i := range weaponsWithTimes {
							m.SelectedWeapons[i] = weaponsWithTimes[i].weapon
							m.RangeInputs[i] = weaponsWithTimes[i].rangeInput
						}
						m.IsSorted = true
						m.SelectionCursor = 0
						m.SelectionScrollOffset = 0
					}
				}
			}
		}

		if m.Focused == "rangeInput" {
			switch msg.String() {
			case "esc", "enter":
				m.Focused = "leftSection"
				if m.CalculationMode == "flightTime" {
					m.RangeInput.Blur()
				} else if m.CalculationMode == "deltaTime" && m.SelectionCursor < len(m.RangeInputs) {
					m.RangeInputs[m.SelectionCursor].Blur()
				}
				return m, nil
			}

			if m.CalculationMode == "flightTime" {
				m.RangeInput, cmd = m.RangeInput.Update(msg)
			} else if m.CalculationMode == "deltaTime" && m.SelectionCursor < len(m.RangeInputs) {
				m.RangeInputs[m.SelectionCursor], cmd = m.RangeInputs[m.SelectionCursor].Update(msg)
			}
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
				fileName := m.FilteredTableRows[m.TableCursor][1]
				for _, weapon := range m.WeaponData {
					if weapon.WeaponFileName == fileName {
						m.SelectedWeapons = append(m.SelectedWeapons, weapon)
						// If this is the first selection, set the range input
						if len(m.SelectedWeapons) == 1 {
							m.SelectionCursor = 0
						}
						// In deltaTime mode, add a new range input for the new weapon
						if m.CalculationMode == "deltaTime" {
							m.RangeInputs = append(m.RangeInputs, textinput.New())
						}
						// Reset sort state when adding weapons
						m.IsSorted = false
						break
					}
				}
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "R":
			m.Focused = "table"
			m.SelectedWeapons = []WeaponData{}
			m.SelectionCursor = 0
			m.SelectionScrollOffset = 0
			m.FilteredTableRows = m.AllTableRows
			m.SearchInput.SetValue("")
			m.RangeInput.SetValue("")
			m.RangeInputs = []textinput.Model{}
			m.TableCursor = 0
			m.TableScrollOffset = 0
			m.IsSorted = false
			m.rebuildTable()
			return m, nil
		case "tab":
			// Blur any focused inputs before switching modes
			if m.Focused == "rangeInput" {
				if m.CalculationMode == "flightTime" {
					m.RangeInput.Blur()
				} else if m.CalculationMode == "deltaTime" && m.SelectionCursor < len(m.RangeInputs) {
					m.RangeInputs[m.SelectionCursor].Blur()
				}
				m.Focused = "leftSection"
			}

			// Toggle calculation mode and reset inputs
			if m.CalculationMode == "flightTime" {
				m.CalculationMode = "deltaTime"
				// Reset and initialize range inputs for all selected weapons
				m.RangeInputs = make([]textinput.Model, len(m.SelectedWeapons))
				for i := range m.RangeInputs {
					m.RangeInputs[i] = textinput.New()
				}
			} else {
				m.CalculationMode = "flightTime"
				// Reset the single range input
				m.RangeInput.SetValue("")
			}
			// Reset sort state when changing modes
			m.IsSorted = false
		case "left", "right", "h", "l":
			if m.Focused == "leftSection" {
				m.Focused = "table"
			} else if m.Focused == "table" {
				m.Focused = "leftSection"
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	m.rebuildTable()

	return m, nil
}

func (m Model) View() string {
	helpText := "↑/↓ k/j: Navigate • ←/→ h/l: Switch section • g/G: Top/Bottom • Tab: Switch mode" + "(" + m.CalculationMode + ")"

	if m.Focused == "table" {
		helpText += " • /: Search • Enter: Add"
		if len(m.FilteredTableRows) > 0 {
			helpText += fmt.Sprintf(" • Row %d/%d", m.TableCursor+1, len(m.FilteredTableRows))
		}
	} else if m.Focused == "leftSection" {
		helpText += " • /: Edit Range • s: Sort • Del: Remove"
		if len(m.SelectedWeapons) > 0 {
			helpText += fmt.Sprintf(" • Selection %d/%d", m.SelectionCursor+1, len(m.SelectedWeapons))
		}
	} else if m.Focused == "rangeInput" {
		helpText += " • Enter/Esc: Done"
	}

	if m.Focused != "rangeInput" {
		helpText += " • R: Reset • q: Quit"
	}

	helpBar := lipgloss.NewStyle().
		Width(m.Width - 2).
		Foreground(lipgloss.Color("245")).
		Background(lipgloss.Color("235")).
		PaddingLeft(1).
		Render(helpText)

	var SelectedWeaponsString strings.Builder
	for _, weapon := range m.SelectedWeapons {
		SelectedWeaponsString.WriteString(weapon.WeaponFileName)
	}

	leftContent := m.renderSelections()

	leftBorderColor := lipgloss.Color("240") // Default gray
	if m.Focused == "leftSection" || m.Focused == "rangeInput" {
		leftBorderColor = lipgloss.Color("62") // Blue when focused
	}

	leftSection := lipgloss.NewStyle().
		Width(m.Width/2 - 2).
		Height(m.Height - 2 - lipgloss.Height(helpBar)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor).
		AlignVertical(lipgloss.Top).
		Render(leftContent)

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

	rightBorderColor := lipgloss.Color("240") // Default gray
	if m.Focused == "table" || m.Focused == "search" {
		rightBorderColor = lipgloss.Color("62") // Blue when focused
	}

	rightSection := lipgloss.NewStyle().
		Width(m.Width/2 - 2).
		Height(m.Height - 2 - lipgloss.Height(helpBar)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
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

	searchInput := textinput.New()
	rangeInput := textinput.New()

	m := Model{
		Renderer:              re,
		AllTableRows:          rows,
		FilteredTableRows:     rows,
		TableCursor:           0,
		SearchInput:           searchInput,
		RangeInput:            rangeInput,
		RangeInputs:           []textinput.Model{},
		Focused:               "table",
		WeaponData:            weapons,
		SelectionCursor:       0,
		SelectionScrollOffset: 0,
		CalculationMode:       "flightTime",
		IsSorted:              false,
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
