package output

import (
	"fmt"
	"time"

	"RATFF/shared"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

const IDColumnWidth = 20

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

func PrintClientTable(clients []shared.ClientInfo, t func(string) string, tf func(string, ...interface{}) string) {
	if len(clients) == 0 {
		PrintInfo(t("no_device_connected"))
		return
	}

	PrintInfo(tf("connected_devices", len(clients)))

	tableData := table.New().
		Headers(t("header_id"), t("header_ip"), t("header_hostname"), t("header_os")).
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
					Width(IDColumnWidth).
					MaxWidth(IDColumnWidth).
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
		tableData.Row(
			FormatID(c.ID),
			c.IP,
			c.Hostname,
			c.OSInfo,
		)
	}

	fmt.Println(tableData)
}

func FormatID(id string) string {
	if lipgloss.Width(id) <= IDColumnWidth {
		return id
	}

	// Keep the beginning and append "..."
	r := []rune(id)
	if len(r) <= IDColumnWidth {
		return id
	}

	return string(r[:IDColumnWidth-3]) + "..."
}

func PrintFileTable(currentPath string, files []interface{}, t func(string) string, tf func(string, ...interface{}) string) {
	if len(files) == 0 {
		PrintInfo(t("file_list_empty"))
		return
	}

	PrintInfo(tf("file_list_path", currentPath))

	tableData := table.New().
		Headers(t("file_header_type"), t("file_header_name"), t("file_header_size"), t("file_header_modified"), t("file_header_permissions")).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle.
					Padding(0, 1).
					Align(lipgloss.Center)
			}

			var s lipgloss.Style
			if row%2 == 0 {
				s = evenRowStyle
			} else {
				s = oddRowStyle
			}

			s = s.Padding(0, 1)

			switch col {
			case 0:
				s = s.Width(10).MaxWidth(10).Align(lipgloss.Center)
			case 1:
				s = s.Align(lipgloss.Left)
			case 2:
				s = s.Width(12).MaxWidth(12).Align(lipgloss.Right)
			case 3:
				s = s.Width(20).MaxWidth(20).Align(lipgloss.Center)
			case 4:
				s = s.Width(12).MaxWidth(12).Align(lipgloss.Center)
			default:
				s = s.Align(lipgloss.Left)
			}

			return s
		})

	for _, f := range files {
		fileMap, ok := f.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := fileMap["name"].(string)
		fileType, _ := fileMap["type"].(string)
		sizeF, _ := fileMap["size"].(float64)
		modTimeF, _ := fileMap["mod_time"].(float64)
		permissions, _ := fileMap["permissions"].(string)
		hidden, _ := fileMap["hidden"].(bool)
		linkTarget, _ := fileMap["link_target"].(string)

		typeIcon := getFileTypeIcon(fileType)
		displayName := name
		if hidden {
			displayName = t("file_hidden_prefix") + name
		}
		if linkTarget != "" {
			displayName = name + " -> " + linkTarget
		}

		sizeStr := formatFileSize(int64(sizeF))
		modTimeStr := formatModTime(int64(modTimeF))

		tableData.Row(
			typeIcon,
			displayName,
			sizeStr,
			modTimeStr,
			permissions,
		)
	}

	fmt.Println(tableData)
}

func getFileTypeIcon(fileType string) string {
	switch fileType {
	case "directory":
		return "📁"
	case "symlink":
		return "🔗"
	case "shortcut":
		return "🔗"
	default:
		return "📄"
	}
}

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.1f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func formatModTime(unixTime int64) string {
	if unixTime == 0 {
		return "N/A"
	}
	t := time.Unix(unixTime, 0)
	return t.Format("2006-01-02 15:04:05")
}
