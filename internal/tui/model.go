package tui

import (
        "fmt"
        "informant/internal/feed"
        "informant/internal/storage"
        "strings"

        tea "github.com/charmbracelet/bubbletea"
)

// ViewMode represents the current view in the TUI
type ViewMode int

const (
        ViewList ViewMode = iota
        ViewReader
        ViewHelp
)

// Model represents the TUI model
type Model struct {
        items       []feed.Item
        storage     *storage.Storage
        viewMode    ViewMode
        cursor      int
        selectedItem *feed.Item
        width       int
        height      int
        scrollOffset int
        showHelp    bool
        err         error
}

// NewModel creates a new TUI model
func NewModel(items []feed.Item, storage *storage.Storage) Model {
        return Model{
                items:    items,
                storage:  storage,
                viewMode: ViewList,
                cursor:   0,
        }
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
        return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        switch msg := msg.(type) {
        case tea.WindowSizeMsg:
                m.width = msg.Width
                m.height = msg.Height

        case tea.KeyMsg:
                switch m.viewMode {
                case ViewList:
                        return m.updateListView(msg)
                case ViewReader:
                        return m.updateReaderView(msg)
                case ViewHelp:
                        return m.updateHelpView(msg)
                }
        }

        return m, nil
}

// updateListView handles key events in list view
func (m Model) updateListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch msg.String() {
        case "q", "ctrl+c":
                return m, tea.Quit

        case "?":
                m.viewMode = ViewHelp
                return m, nil

        case "j", "down":
                if m.cursor < len(m.items)-1 {
                        m.cursor++
                        m.adjustScroll()
                }

        case "k", "up":
                if m.cursor > 0 {
                        m.cursor--
                        m.adjustScroll()
                }

        case "g":
                m.cursor = 0
                m.scrollOffset = 0

        case "G":
                m.cursor = len(m.items) - 1
                m.adjustScroll()

        case "enter":
                if len(m.items) > 0 {
                        m.selectedItem = &m.items[m.cursor]
                        m.viewMode = ViewReader
                }

        case "r":
                // Toggle read status
                if len(m.items) > 0 {
                        item := &m.items[m.cursor]
                        if m.storage.IsRead(item.ID) {
                                m.storage.MarkAsUnread(item.ID)
                        } else {
                                m.storage.MarkAsRead(item.ID)
                        }
                }
        }

        return m, nil
}

// updateReaderView handles key events in reader view
func (m Model) updateReaderView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch msg.String() {
        case "q", "escape":
                m.viewMode = ViewList
                m.selectedItem = nil

        case "r":
                // Toggle read status of current item
                if m.selectedItem != nil {
                        if m.storage.IsRead(m.selectedItem.ID) {
                                m.storage.MarkAsUnread(m.selectedItem.ID)
                        } else {
                                m.storage.MarkAsRead(m.selectedItem.ID)
                        }
                }

        case "j", "down":
                // Scroll content down
                m.scrollOffset++

        case "k", "up":
                // Scroll content up
                if m.scrollOffset > 0 {
                        m.scrollOffset--
                }
        }

        return m, nil
}

// updateHelpView handles key events in help view
func (m Model) updateHelpView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch msg.String() {
        case "q", "escape", "?":
                m.viewMode = ViewList
        }

        return m, nil
}

// adjustScroll adjusts scroll offset to keep cursor visible
func (m *Model) adjustScroll() {
        visibleHeight := m.height - 4 // Account for header and status
        
        if m.cursor < m.scrollOffset {
                m.scrollOffset = m.cursor
        } else if m.cursor >= m.scrollOffset+visibleHeight {
                m.scrollOffset = m.cursor - visibleHeight + 1
        }
}

// View renders the current view
func (m Model) View() string {
        if m.width == 0 {
                return "Loading..."
        }

        switch m.viewMode {
        case ViewList:
                return m.renderListView()
        case ViewReader:
                return m.renderReaderView()
        case ViewHelp:
                return m.renderHelpView()
        default:
                return "Unknown view"
        }
}

// renderListView renders the list of news items
func (m Model) renderListView() string {
        var b strings.Builder

        // Header
        header := headerStyle.Render("Informant - Arch Linux News Reader")
        b.WriteString(header + "\n")

        // Status line
        unreadCount := 0
        for _, item := range m.items {
                if !m.storage.IsRead(item.ID) {
                        unreadCount++
                }
        }

        status := fmt.Sprintf("Items: %d | Unread: %d | Use ? for help", len(m.items), unreadCount)
        b.WriteString(statusStyle.Render(status) + "\n\n")

        // Items list
        visibleHeight := m.height - 6 // Account for header, status, and help
        start := m.scrollOffset
        end := start + visibleHeight

        if end > len(m.items) {
                end = len(m.items)
        }

        for i := start; i < end; i++ {
                item := m.items[i]
                isSelected := (i == m.cursor)
                isRead := m.storage.IsRead(item.ID)

                // Format item line
                status := "●"
                if isRead {
                        status = "○"
                }

                // Format date
                dateStr := item.Published.Format("2006-01-02")

                feedInfo := ""
                if item.FeedName != "" {
                        feedInfo = fmt.Sprintf(" (%s)", item.FeedName)
                }

                line := fmt.Sprintf("%s %s %s%s", status, dateStr, item.Title, feedInfo)

                // Truncate if too long
                maxWidth := m.width - 4
                if len(line) > maxWidth {
                        line = line[:maxWidth-3] + "..."
                }

                // Apply style
                style := GetItemStyle(isSelected, isRead)
                if isSelected {
                        line = "▶ " + line
                } else {
                        line = "  " + line
                }

                b.WriteString(style.Render(line) + "\n")
        }

        // Scroll indicator
        if len(m.items) > visibleHeight {
                scrollInfo := fmt.Sprintf("[%d/%d]", m.cursor+1, len(m.items))
                b.WriteString("\n" + statusStyle.Render(scrollInfo))
        }

        // Help hint
        b.WriteString("\n" + helpStyle.Render("Press ? for help, q to quit"))

        return b.String()
}

// renderReaderView renders the content of a selected item
func (m Model) renderReaderView() string {
        if m.selectedItem == nil {
                return errorStyle.Render("No item selected")
        }

        var b strings.Builder

        // Header with title
        title := fmt.Sprintf("Reading: %s", m.selectedItem.Title)
        header := contentHeaderStyle.Render(title)
        b.WriteString(header + "\n")

        // Meta information
        dateStr := m.selectedItem.Published.Format("2006-01-02 15:04:05")
        meta := dateStyle.Render("Date: " + dateStr)
        
        if m.selectedItem.FeedName != "" {
                meta += " | " + feedNameStyle.Render("Feed: "+m.selectedItem.FeedName)
        }

        readStatus := "Unread"
        if m.storage.IsRead(m.selectedItem.ID) {
                readStatus = "Read"
        }
        meta += " | Status: " + readStatus

        b.WriteString(meta + "\n\n")

        // Content with scroll
        content := m.selectedItem.Content
        lines := strings.Split(content, "\n")
        
        visibleHeight := m.height - 8 // Account for header, meta, and controls
        start := m.scrollOffset
        end := start + visibleHeight

        if end > len(lines) {
                end = len(lines)
        }

        if start < len(lines) {
                visibleContent := strings.Join(lines[start:end], "\n")
                b.WriteString(contentStyle.Width(m.width-4).Render(visibleContent))
        }

        // Scroll indicator
        if len(lines) > visibleHeight {
                scrollInfo := fmt.Sprintf("[Line %d-%d of %d]", start+1, end, len(lines))
                b.WriteString("\n" + statusStyle.Render(scrollInfo))
        }

        // Controls
        b.WriteString("\n" + helpStyle.Render("j/k: scroll | r: toggle read | q: back to list"))

        return b.String()
}

// renderHelpView renders the help screen
func (m Model) renderHelpView() string {
        var b strings.Builder

        header := contentHeaderStyle.Render("Informant Help")
        b.WriteString(header + "\n\n")

        helpText := [][]string{
                {"Navigation", ""},
                {"j, ↓", "Move down"},
                {"k, ↑", "Move up"},
                {"g", "Go to first item"},
                {"G", "Go to last item"},
                {"", ""},
                {"Actions", ""},
                {"Enter", "Read selected item"},
                {"r", "Toggle read/unread status"},
                {"?", "Show/hide this help"},
                {"q", "Quit application"},
                {"", ""},
                {"Reader Mode", ""},
                {"j, ↓", "Scroll content down"},
                {"k, ↑", "Scroll content up"},
                {"r", "Toggle read status"},
                {"q, Esc", "Back to list"},
        }

        for _, row := range helpText {
                if row[0] == "" {
                        b.WriteString("\n")
                        continue
                }

                if row[1] == "" {
                        // Section header
                        b.WriteString(titleStyle.Render(row[0]) + "\n")
                } else {
                        // Key binding
                        key := helpKeyStyle.Render(row[0])
                        desc := helpStyle.Render(row[1])
                        line := fmt.Sprintf("  %-12s %s", key, desc)
                        b.WriteString(line + "\n")
                }
        }

        b.WriteString("\n" + helpStyle.Render("Press ? or q to close help"))

        return contentStyle.Width(m.width-4).Render(b.String())
}
