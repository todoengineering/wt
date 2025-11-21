package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle      = lipgloss.NewStyle().MarginLeft(2)
	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

// Item represents a list item in the selector
type Item struct {
	TitleStr       string
	DescriptionStr string
	FilterStr      string
	// Value is the underlying data for the item
	Value interface{}
}

func (i Item) Title() string       { return i.TitleStr }
func (i Item) Description() string { return i.DescriptionStr }
func (i Item) FilterValue() string { return i.FilterStr }

type model struct {
	list     list.Model
	choice   *Item
	quitting bool
	err      error
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(Item)
			if ok {
				m.choice = &i
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != nil {
		return ""
	}
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

// Select displays a list of items and allows the user to select one.
// It returns the selected Item and any error that occurred.
func Select(items []Item, title string) (Item, error) {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	const defaultWidth = 20
	const listHeight = 14

	l := list.New(listItems, list.NewDefaultDelegate(), defaultWidth, listHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{list: l}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return Item{}, fmt.Errorf("error running selector: %w", err)
	}

	finalM := finalModel.(model)
	if finalM.choice != nil {
		return *finalM.choice, nil
	}

	if finalM.err != nil {
		return Item{}, finalM.err
	}

	return Item{}, fmt.Errorf("selection cancelled")
}
