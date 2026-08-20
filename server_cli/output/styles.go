package output

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	styleDebug   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	stylePrompt  = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))

	progressBarStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E"))
	progressPercentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")).Bold(true)
	progressSpeedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24"))
	progressETAStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA"))
	progressFilenameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB")).Bold(true)
)

func PrintSuccess(msg string) {
	prefix := styleSuccess.Render("[+]")
	fmt.Printf("%s %s\n", prefix, msg)
}

func PrintError(msg string) {
	prefix := styleError.Render("[-]")
	fmt.Printf("%s %s\n", prefix, msg)
}

func PrintInfo(msg string) {
	prefix := styleInfo.Render("[*]")
	fmt.Printf("%s %s\n", prefix, msg)
}

func PrintDebug(msg string) {
	prefix := styleDebug.Render("[debug]")
	fmt.Printf("%s %s\n", prefix, msg)
}

func PrintWarn(msg string) {
	prefix := styleWarn.Render("[!]")
	fmt.Printf("%s %s\n", prefix, msg)
}

func BuildPrompt(id string, inCommandMode bool) string {
	if id == "" {
		return stylePrompt.Render("(server) >> ")
	}
	if inCommandMode {
		return stylePrompt.Render(fmt.Sprintf("(%s)(command) >> ", id))
	}
	return stylePrompt.Render(fmt.Sprintf("(%s)(console) >> ", id))
}

func StyleCommandOutput(outputStr string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 2).
		Render(outputStr)
}

func PrintCommandResult(stdout, stderr string, exitCode int, t func(string) string, tf func(string, ...interface{}) string) {
	if stdout != "" {
		PrintInfo(t("command_stdout"))
		fmt.Println(StyleCommandOutput(stdout))
		fmt.Println()
	}

	if stderr != "" {
		PrintError(t("command_stderr"))
		fmt.Println(StyleCommandOutput(stderr))
		fmt.Println()
	}

	if exitCode == 0 {
		PrintSuccess(tf("command_exit_code", exitCode))
	} else {
		PrintError(tf("command_exit_code", exitCode))
	}
}

type ProgressBar struct {
	mu        sync.Mutex
	total     int64
	current   int64
	width     int
	startTime time.Time
	filename  string
	done      bool
}

func NewProgressBar(total int64, filename string) *ProgressBar {
	return &ProgressBar{
		total:     total,
		width:     30,
		startTime: time.Now(),
		filename:  filename,
	}
}

func (p *ProgressBar) Add(n int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current += n
}

func (p *ProgressBar) SetTotal(total int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total = total
}

func (p *ProgressBar) MarkDone() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done = true
}

func (p *ProgressBar) Display() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.total == 0 {
		return
	}

	percent := float64(p.current) / float64(p.total) * 100
	if percent > 100 {
		percent = 100
	}

	filled := int(percent / 100 * float64(p.width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)
	barStyled := progressBarStyle.Render(bar)

	elapsed := time.Since(p.startTime).Seconds()
	speed := float64(p.current) / 1024 / 1024 / elapsed

	remaining := float64(p.total-p.current) / (float64(p.current) / elapsed)
	eta := time.Duration(remaining * float64(time.Second))

	percentStr := progressPercentStyle.Render(fmt.Sprintf("%.1f%%", percent))
	speedStr := progressSpeedStyle.Render(fmt.Sprintf("%.2f MB/s", speed))
	etaStr := progressETAStyle.Render(fmt.Sprintf("ETA: %s", FormatDuration(eta)))
	filenameStr := progressFilenameStyle.Render(p.filename)

	line := fmt.Sprintf("\r[%s] %s %s | %s/%s | %s | %s",
		barStyled,
		filenameStr,
		percentStr,
		FormatBytes(p.current),
		FormatBytes(p.total),
		speedStr,
		etaStr,
	)

	if p.done {
		fmt.Println()
		return
	}

	fmt.Print(line)
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

var (
	sysinfoTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#8B5CF6")).
				MarginBottom(1)

	sysinfoSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#60A5FA")).
				MarginTop(1).
				MarginBottom(1)

	sysinfoKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			Width(18).
			Align(lipgloss.Right)

	sysinfoValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5E7EB"))

	sysinfoErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444"))

	IpInfoCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#60A5FA")).
			Padding(1, 2).
			Width(50)

	IpInfoTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#8B5CF6")).
				MarginBottom(1)

	IpInfoKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			Width(16).
			Align(lipgloss.Right)

	IpInfoValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5E7EB"))

	IpInfoErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				Bold(true)
)

// PrintIPInfoCard prints public IP information in a styled card.
func PrintIPInfoCard(ip, continent, country, countryCode, region, city, isp, timezone string, lat, lon float64) {
	var lines []string

	lines = append(lines, IpInfoTitleStyle.Render("Public IP Information"))
	lines = append(lines, "")

	if ip != "" {
		lines = append(lines, formatIPInfoLine("IP:", ip))
	}
	if continent != "" {
		lines = append(lines, formatIPInfoLine("Continent:", continent))
	}
	if country != "" {
		countryText := country
		if countryCode != "" {
			countryText += fmt.Sprintf(" (%s)", countryCode)
		}
		lines = append(lines, formatIPInfoLine("Country:", countryText))
	}
	if region != "" {
		lines = append(lines, formatIPInfoLine("Region:", region))
	}
	if city != "" {
		lines = append(lines, formatIPInfoLine("City:", city))
	}
	if isp != "" {
		lines = append(lines, formatIPInfoLine("ISP:", isp))
	}
	if timezone != "" {
		lines = append(lines, formatIPInfoLine("Timezone:", timezone))
	}
	if lat != 0 && lon != 0 {
		lines = append(lines, formatIPInfoLine("Location:", fmt.Sprintf("%.4f, %.4f", lat, lon)))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	fmt.Println(IpInfoCardStyle.Render(content))
	fmt.Println()
}

func formatIPInfoLine(key, value string) string {
	keyStr := IpInfoKeyStyle.Render(key)
	valueStr := IpInfoValueStyle.Render(value)
	return lipgloss.JoinHorizontal(lipgloss.Top, "  ", keyStr, " ", valueStr)
}
