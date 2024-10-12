package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UIModel struct {
	items       []string
	totalItems  int
	currentItem int
	itemStyle   lipgloss.Style
	pb          progress.Model
}

func NewUIModel(totalItems int) *UIModel {
	return &UIModel{
		items:      make([]string, 0, 10),
		totalItems: totalItems,
		itemStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")),
		pb:         progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C")),
	}
}

func (m UIModel) Init() tea.Cmd {
	return nil
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case DownloadEvent:
		item := fmt.Sprintf("%s: %s", msg.Collection, msg.Icon)
		if len(m.items) < 10 {
			m.items = append(m.items, item)
		} else {
			m.items = append(m.items[1:], item)
		}
		m.currentItem++
	}

	return m, nil
}

func (m UIModel) View() string {
	s := "Dowloading Icons\n"
	s += fmt.Sprintf("current: %d / total: %d\n\n", m.currentItem, m.totalItems)
	for _, item := range m.items {
		s += m.itemStyle.Render(item) + "\n"
	}

	s += "\n"
	s += m.pb.ViewAs((float64(m.currentItem) * 100 / float64(m.totalItems) / 100))
	s += "\n"
	return s
}

type DownloadEvent struct {
	Collection string
	Icon       string
}
