package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color scheme
	primaryColor   = lipgloss.Color("12")  // Blue
	secondaryColor = lipgloss.Color("8")   // Gray
	accentColor    = lipgloss.Color("10")  // Green
	warningColor   = lipgloss.Color("11")  // Yellow
	errorColor     = lipgloss.Color("9")   // Red

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// List item styles
	itemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Background(primaryColor).
				Foreground(lipgloss.Color("0")).
				Padding(0, 1)

	unreadItemStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true).
			Padding(0, 1)

	selectedUnreadItemStyle = lipgloss.NewStyle().
				Background(warningColor).
				Foreground(lipgloss.Color("0")).
				Bold(true).
				Padding(0, 1)

	readItemStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Padding(0, 1)

	selectedReadItemStyle = lipgloss.NewStyle().
				Background(secondaryColor).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)

	// Content styles
	contentStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor)

	contentHeaderStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Margin(0, 0, 1, 0)

	feedNameStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Italic(true)

	dateStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Help styles
	helpStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Padding(1, 0)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Status styles
	statusStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Pager styles
	pagerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	pagerHeaderStyle = lipgloss.NewStyle().
				Background(primaryColor).
				Foreground(lipgloss.Color("15")).
				Bold(true).
				Padding(0, 1).
				Margin(0, 0, 1, 0)

	// Viewport scrollbar
	scrollbarThumbStyle = lipgloss.NewStyle().
				Background(primaryColor)

	scrollbarTrackStyle = lipgloss.NewStyle().
				Background(secondaryColor)
)

// GetItemStyle returns the appropriate style for a list item
func GetItemStyle(isSelected, isRead bool) lipgloss.Style {
	switch {
	case isSelected && !isRead:
		return selectedUnreadItemStyle
	case isSelected && isRead:
		return selectedReadItemStyle
	case !isSelected && !isRead:
		return unreadItemStyle
	default:
		return readItemStyle
	}
}
