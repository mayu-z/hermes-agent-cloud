package ui

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hermes-tui/client"
)

// ── State machine ─────────────────────────────────────────────────────────────

type appState int

const (
	stateBooting appState = iota
	stateChat
	stateTasks
	stateClock
	stateExplorer
	stateHelp
)

// ── Tea messages ──────────────────────────────────────────────────────────────

type bootTickMsg struct{}
type chatResponseMsg struct {
	resp *client.ChatResponse
	err  error
}
type profileMsg struct {
	profile   *client.Profile
	taskCount int
	connected bool
}
type tasksMsg struct{ tasks []client.Task }
type tickMsg time.Time

// ── Chat message ──────────────────────────────────────────────────────────────

type ChatMsg struct {
	Role    string // user | hermes | tool | error | system
	Content string
	Time    time.Time
}

// ── Explorer ──────────────────────────────────────────────────────────────────

type explorerState struct {
	dir     string
	entries []os.DirEntry
	cursor  int
	err     string
}

// ── Model ─────────────────────────────────────────────────────────────────────

type Model struct {
	client    *client.Client
	state     appState
	msgs      []ChatMsg
	input     string
	thinking  bool
	width     int
	height    int
	scroll    int
	profile   *client.Profile
	taskCount int
	connected bool
	tasks     []client.Task

	// boot
	bootStep int
	bootDone bool

	// clock
	showClock bool
	now       time.Time

	// explorer
	ex explorerState

	// thinking animation
	thinkFrame int
}

func NewModel(c *client.Client) Model {
	home, _ := os.UserHomeDir()
	return Model{
		client: c,
		state:  stateBooting,
		now:    time.Now(),
		ex:     explorerState{dir: home},
	}
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		bootTick(),
		tickEvery(),
	)
}

func bootTick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return bootTickMsg{}
	})
}

func tickEvery() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		m.now = time.Time(msg)
		if m.thinking {
			m.thinkFrame = (m.thinkFrame + 1) % 4
		}
		return m, tickEvery()

	case bootTickMsg:
		if m.bootStep < len(BootLines) {
			m.bootStep++
			return m, bootTick()
		}
		// boot done — load profile
		m.state = stateChat
		return m, m.cmdLoadProfile()

	case profileMsg:
		m.profile = msg.profile
		m.taskCount = msg.taskCount
		m.connected = msg.connected
		return m, nil

	case tasksMsg:
		m.tasks = msg.tasks
		return m, nil

	case chatResponseMsg:
		m.thinking = false
		if msg.err != nil {
			m.msgs = append(m.msgs, ChatMsg{Role: "error", Content: msg.err.Error(), Time: time.Now()})
		} else {
			if len(msg.resp.ToolsUsed) > 0 {
				m.msgs = append(m.msgs, ChatMsg{
					Role:    "tool",
					Content: fmtTools(msg.resp.ToolsUsed),
					Time:    time.Now(),
				})
			}
			m.msgs = append(m.msgs, ChatMsg{Role: "hermes", Content: msg.resp.Reply, Time: time.Now()})
			if msg.resp.Profile != nil {
				m.profile = msg.resp.Profile
			}
		}
		m.scroll = 0
		return m, m.cmdLoadProfile()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global shortcuts
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyCtrlP:
		m.showClock = !m.showClock
		return m, nil

	case tea.KeyCtrlK:
		if m.state == stateHelp {
			m.state = stateChat
		} else {
			m.state = stateHelp
		}
		return m, nil

	case tea.KeyCtrlH:
		if m.state == stateExplorer {
			m.state = stateChat
		} else {
			m.state = stateExplorer
			m.ex.cursor = 0
			m.ex.err = ""
			entries, err := os.ReadDir(m.ex.dir)
			if err != nil {
				m.ex.err = err.Error()
			} else {
				m.ex.entries = entries
			}
		}
		return m, nil

	case tea.KeyCtrlL:
		m.client.ClearHistory()
		m.msgs = []ChatMsg{}
		m.scroll = 0
		m.msgs = append(m.msgs, ChatMsg{Role: "system", Content: "-- session cleared --", Time: time.Now()})
		return m, nil

	case tea.KeyCtrlT:
		if m.state == stateTasks {
			m.state = stateChat
		} else {
			m.state = stateTasks
			return m, m.cmdLoadTasks()
		}
		return m, nil
	}

	// State-specific handling
	switch m.state {
	case stateExplorer:
		return m.handleExplorer(msg)
	case stateChat:
		return m.handleChat(msg)
	}

	return m, nil
}

func (m Model) handleChat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if m.thinking || strings.TrimSpace(m.input) == "" {
			return m, nil
		}
		text := strings.TrimSpace(m.input)
		m.msgs = append(m.msgs, ChatMsg{Role: "user", Content: text, Time: time.Now()})
		m.input = ""
		m.thinking = true
		m.scroll = 0
		return m, m.cmdChat(text)

	case tea.KeyBackspace:
		if len(m.input) > 0 {
			runes := []rune(m.input)
			m.input = string(runes[:len(runes)-1])
		}

	case tea.KeyUp:
		m.scroll++
	case tea.KeyDown:
		if m.scroll > 0 {
			m.scroll--
		}

	case tea.KeySpace:
		if !m.thinking {
			m.input += " "
		}

	case tea.KeyRunes:
		if !m.thinking {
			m.input += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleExplorer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	entries := m.ex.entries
	switch msg.Type {
	case tea.KeyUp:
		if m.ex.cursor > 0 {
			m.ex.cursor--
		}
	case tea.KeyDown:
		if m.ex.cursor < len(entries)-1 {
			m.ex.cursor++
		}
	case tea.KeyEnter:
		if len(entries) == 0 {
			return m, nil
		}
		entry := entries[m.ex.cursor]
		full := filepath.Join(m.ex.dir, entry.Name())
		if entry.IsDir() {
			// navigate into dir
			newEntries, err := os.ReadDir(full)
			if err != nil {
				m.ex.err = err.Error()
			} else {
				m.ex.dir = full
				m.ex.entries = newEntries
				m.ex.cursor = 0
				m.ex.err = ""
			}
		} else {
			// send file to hermes
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext == ".pdf" {
				data, err := os.ReadFile(full)
				if err != nil {
					m.ex.err = "cannot read file: " + err.Error()
					return m, nil
				}
				b64 := base64.StdEncoding.EncodeToString(data)
				msg := fmt.Sprintf("[PDF attached: %s]\n\nAnalyze this PDF for me. pdf_base64=%s", entry.Name(), b64)
				m.msgs = append(m.msgs, ChatMsg{Role: "system", Content: fmt.Sprintf("📄 Uploading %s...", entry.Name()), Time: time.Now()})
				m.thinking = true
				m.state = stateChat
				m.scroll = 0
				return m, m.cmdChat(msg)
			} else if ext == ".txt" || ext == ".md" {
				data, err := os.ReadFile(full)
				if err != nil {
					m.ex.err = "cannot read file: " + err.Error()
					return m, nil
				}
				msg := fmt.Sprintf("[File: %s]\n\n%s", entry.Name(), string(data))
				m.msgs = append(m.msgs, ChatMsg{Role: "system", Content: fmt.Sprintf("📄 Sending %s...", entry.Name()), Time: time.Now()})
				m.thinking = true
				m.state = stateChat
				return m, m.cmdChat(msg)
			} else {
				m.ex.err = fmt.Sprintf("unsupported type: %s (supported: .pdf .txt .md)", ext)
			}
		}
	case tea.KeyBackspace:
		// go up one directory
		parent := filepath.Dir(m.ex.dir)
		if parent != m.ex.dir {
			newEntries, err := os.ReadDir(parent)
			if err != nil {
				m.ex.err = err.Error()
			} else {
				m.ex.dir = parent
				m.ex.entries = newEntries
				m.ex.cursor = 0
				m.ex.err = ""
			}
		}
	}
	return m, nil
}

// ── Commands ──────────────────────────────────────────────────────────────────

func (m Model) cmdChat(text string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.Chat(text)
		return chatResponseMsg{resp, err}
	}
}

func (m Model) cmdLoadProfile() tea.Cmd {
	return func() tea.Msg {
		r, err := m.client.GetProfile()
		if err != nil {
			return profileMsg{nil, 0, false}
		}
		count := 0
		var prof *client.Profile
		if r != nil {
			prof = r.Profile
			count = r.Tasks.ActiveCount
		}
		return profileMsg{prof, count, true}
	}
}

func (m Model) cmdLoadTasks() tea.Cmd {
	return func() tea.Msg {
		r, err := m.client.GetTasks()
		if err != nil || r == nil {
			return tasksMsg{nil}
		}
		return tasksMsg{r.Tasks}
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	// Boot sequence
	if m.state == stateBooting {
		return m.viewBoot()
	}

	sideW := 26
	chatW := m.width - sideW - 1
	if chatW < 30 {
		chatW = m.width
		sideW = 0
	}

	var main string
	switch m.state {
	case stateTasks:
		main = m.viewTasks(chatW)
	case stateExplorer:
		main = m.viewExplorer(chatW)
	case stateHelp:
		main = m.viewHelp(chatW)
	default:
		main = m.viewChat(chatW)
	}

	var out string
	if sideW > 0 {
		sidebar := RenderSidebar(m.profile, m.taskCount, m.client.SessionID, m.connected, m.height)
		out = lipgloss.JoinHorizontal(lipgloss.Top, main, sidebar)
	} else {
		out = main
	}

	// Clock overlay (just appended as a floating line at bottom)
	if m.showClock {
		clock := ClockStyle.Render(m.now.Format("  15:04:05  Mon 02 Jan 2006  "))
		out = out + "\n" + clock
	}

	return out
}

// ── Boot view ─────────────────────────────────────────────────────────────────

func (m Model) viewBoot() string {
	var sb strings.Builder

	ascii := BootStyle.Render(HermesASCII)
	sb.WriteString(ascii)
	sb.WriteString("\n")

	for i := 0; i < m.bootStep && i < len(BootLines); i++ {
		line := BootLines[i]
		if strings.HasPrefix(line, "[OK]") {
			sb.WriteString(BootStyle.Render(line))
		} else if strings.HasPrefix(line, "━") {
			sb.WriteString(DividerStyle.Render(line))
		} else {
			sb.WriteString(BootDimStyle.Render(line))
		}
		sb.WriteString("\n")
	}

	// blinking cursor at end of boot
	if m.bootStep >= len(BootLines) {
		sb.WriteString(PromptStyle.Render("▊"))
	}

	return sb.String()
}

// ── Chat view ─────────────────────────────────────────────────────────────────

func (m Model) viewChat(width int) string {
	var sb strings.Builder

	// Header
	header := fmt.Sprintf(" HERMES // %s ", m.now.Format("15:04"))
	headerLine := BootStyle.Width(width).Render(header)
	sb.WriteString(headerLine)
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", width)))
	sb.WriteString("\n")

	// Messages area
	inputH := 2
	statusH := 1
	headerH := 2
	msgH := m.height - inputH - statusH - headerH - 1

	rendered := m.renderMsgs(width - 2)
	lines := strings.Split(rendered, "\n")

	start := 0
	if len(lines) > msgH {
		start = len(lines) - msgH - m.scroll
		if start < 0 {
			start = 0
		}
	}
	end := start + msgH
	if end > len(lines) {
		end = len(lines)
	}

	visible := lines[start:end]
	sb.WriteString(strings.Join(visible, "\n"))

	// Pad
	for i := len(visible); i < msgH; i++ {
		sb.WriteString("\n")
	}

	sb.WriteString(DividerStyle.Render(strings.Repeat("─", width)))
	sb.WriteString("\n")

	// Input line
	if m.thinking {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinner := frames[m.thinkFrame%len(frames)]
		sb.WriteString(ToolStyle.Render(fmt.Sprintf(" %s hermes is thinking...", spinner)))
	} else {
		prefix := InputPrefixStyle.Render(" you@hermes:~$ ")
		cursor := InputStyle.Render(m.input + "▊")
		sb.WriteString(prefix + cursor)
	}

	sb.WriteString("\n")

	// Status bar
	status := fmt.Sprintf(" Ctrl+K:help  Ctrl+H:files  Ctrl+T:tasks  Ctrl+P:clock  Ctrl+L:clear  Ctrl+C:quit ")
	sb.WriteString(StatusStyle.Width(width).Render(status))

	return sb.String()
}

func (m Model) renderMsgs(width int) string {
	if len(m.msgs) == 0 {
		return HermesMsgStyle.Render("\n hermes@agent:~$ _\n Hello. I am Hermes. Your AI agent is ready.\n Tell me your name, ask me anything, or give me a task.\n")
	}

	var sb strings.Builder
	for _, msg := range m.msgs {
		switch msg.Role {
		case "user":
			sb.WriteString(UserPrefixStyle.Render(" you  "))
			sb.WriteString(DividerStyle.Render(msg.Time.Format("15:04:05")))
			sb.WriteString("\n")
			for _, line := range wrapLines(msg.Content, width-3) {
				sb.WriteString(UserMsgStyle.Render(" > " + line))
				sb.WriteString("\n")
			}
			sb.WriteString("\n")

		case "hermes":
			sb.WriteString(HermesPrefixStyle.Render(" hermes "))
			sb.WriteString(DividerStyle.Render(msg.Time.Format("15:04:05")))
			sb.WriteString("\n")
			for _, line := range wrapLines(msg.Content, width-3) {
				sb.WriteString(HermesMsgStyle.Render(" | " + line))
				sb.WriteString("\n")
			}
			sb.WriteString("\n")

		case "tool":
			sb.WriteString(ToolStyle.Render(" ⚙ " + msg.Content))
			sb.WriteString("\n")

		case "error":
			sb.WriteString(ErrorStyle.Render(" ✗ ERROR: " + msg.Content))
			sb.WriteString("\n\n")

		case "system":
			sb.WriteString(DividerStyle.Render(" -- " + msg.Content + " --"))
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// ── Tasks view ────────────────────────────────────────────────────────────────

func (m Model) viewTasks(width int) string {
	var sb strings.Builder
	sb.WriteString(BootStyle.Render(" TASK LIST "))
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", width)))
	sb.WriteString("\n\n")

	active := []client.Task{}
	done := []client.Task{}
	for _, t := range m.tasks {
		if t.Completed {
			done = append(done, t)
		} else {
			active = append(active, t)
		}
	}

	if len(active) == 0 && len(done) == 0 {
		sb.WriteString(HermesMsgStyle.Render(" No tasks yet. Tell Hermes to add some.\n"))
	}

	if len(active) > 0 {
		sb.WriteString(SidebarLabelStyle.Render(" ACTIVE:\n"))
		for _, t := range active {
			var ps lipgloss.Style
			switch t.Priority {
			case "high":
				ps = PriorityHigh
			case "low":
				ps = PriorityLow
			default:
				ps = PriorityMed
			}
			sb.WriteString(ps.Render(fmt.Sprintf("  [%s] %s\n", strings.ToUpper(t.Priority), t.TaskDesc)))
		}
		sb.WriteString("\n")
	}

	if len(done) > 0 {
		sb.WriteString(SidebarLabelStyle.Render(" DONE:\n"))
		for _, t := range done {
			sb.WriteString(DividerStyle.Render(fmt.Sprintf("  [✓] %s\n", t.TaskDesc)))
		}
	}

	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", width)))
	sb.WriteString("\n")
	sb.WriteString(StatusStyle.Render(" Ctrl+T to close "))
	return sb.String()
}

// ── Explorer view ─────────────────────────────────────────────────────────────

func (m Model) viewExplorer(width int) string {
	var sb strings.Builder
	sb.WriteString(BootStyle.Render(" FILE EXPLORER "))
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", width)))
	sb.WriteString("\n")
	sb.WriteString(SidebarLabelStyle.Render(" DIR: "))
	sb.WriteString(SidebarValueStyle.Render(m.ex.dir))
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", width)))
	sb.WriteString("\n")

	if m.ex.err != "" {
		sb.WriteString(ErrorStyle.Render(" ✗ " + m.ex.err))
		sb.WriteString("\n")
	}

	maxShow := m.height - 8
	start := 0
	if m.ex.cursor >= maxShow {
		start = m.ex.cursor - maxShow + 1
	}

	for i := start; i < len(m.ex.entries) && i < start+maxShow; i++ {
		entry := m.ex.entries[i]
		prefix := "  "
		if i == m.ex.cursor {
			prefix = "> "
		}

		name := entry.Name()
		if entry.IsDir() {
			name = name + "/"
			if i == m.ex.cursor {
				sb.WriteString(PromptStyle.Render(prefix + name))
			} else {
				sb.WriteString(HermesMsgStyle.Render(prefix + name))
			}
		} else {
			if i == m.ex.cursor {
				sb.WriteString(UserMsgStyle.Render(prefix + name))
			} else {
				sb.WriteString(SidebarValueStyle.Render(prefix + name))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", width)))
	sb.WriteString("\n")
	sb.WriteString(StatusStyle.Render(" ↑↓:navigate  Enter:open/send  Backspace:parent  Ctrl+H:close "))
	return sb.String()
}

// ── Help view ─────────────────────────────────────────────────────────────────

func (m Model) viewHelp(width int) string {
	var sb strings.Builder
	sb.WriteString(BootStyle.Render(" HERMES — KEYBOARD SHORTCUTS "))
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", width)))
	sb.WriteString("\n\n")

	helps := [][2]string{
		{"Enter", "Send message"},
		{"Ctrl+K", "Toggle this help screen"},
		{"Ctrl+L", "Clear chat history"},
		{"Ctrl+T", "Toggle task list"},
		{"Ctrl+H", "Open file explorer (upload PDF/TXT)"},
		{"Ctrl+P", "Toggle clock overlay"},
		{"Ctrl+C", "Quit Hermes"},
		{"↑ / ↓", "Scroll chat history"},
		{"Backspace", "(in explorer) Go to parent directory"},
		{"Enter", "(in explorer) Open folder or send file to Hermes"},
	}

	for _, h := range helps {
		key := PromptStyle.Render(fmt.Sprintf("  %-12s", h[0]))
		desc := HermesMsgStyle.Render(h[1])
		sb.WriteString(key + "  " + desc + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", width)))
	sb.WriteString("\n")
	sb.WriteString(HermesMsgStyle.Render("\n TIPS:\n"))
	sb.WriteString(SidebarValueStyle.Render("  • Tell Hermes your name — it will remember you.\n"))
	sb.WriteString(SidebarValueStyle.Render("  • Ask Hermes to search the web for anything.\n"))
	sb.WriteString(SidebarValueStyle.Render("  • Paste a YouTube URL and ask Hermes to summarize it.\n"))
	sb.WriteString(SidebarValueStyle.Render("  • Open a PDF via Ctrl+H and Hermes will analyze it.\n"))
	sb.WriteString(SidebarValueStyle.Render("  • Say 'add task: ...' to create a task.\n"))
	sb.WriteString("\n")
	sb.WriteString(StatusStyle.Render(" Ctrl+K to close "))
	return sb.String()
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func wrapLines(text string, width int) []string {
	if width <= 0 {
		width = 80
	}
	var result []string
	for _, line := range strings.Split(text, "\n") {
		if len(line) <= width {
			result = append(result, line)
			continue
		}
		words := strings.Fields(line)
		cur := ""
		for _, w := range words {
			if cur == "" {
				cur = w
			} else if len(cur)+1+len(w) <= width {
				cur += " " + w
			} else {
				result = append(result, cur)
				cur = w
			}
		}
		if cur != "" {
			result = append(result, cur)
		}
	}
	if len(result) == 0 {
		result = []string{""}
	}
	return result
}

func fmtTools(tools []string) string {
	names := map[string]string{
		"web_search":        "🔍 web search",
		"youtube_summarize": "📺 youtube",
		"pdf_analyze":       "📄 pdf",
		"save_memory":       "🧠 memory",
		"manage_tasks":      "📋 tasks",
	}
	seen := map[string]bool{}
	var parts []string
	for _, t := range tools {
		if seen[t] {
			continue
		}
		seen[t] = true
		if n, ok := names[t]; ok {
			parts = append(parts, n)
		} else {
			parts = append(parts, t)
		}
	}
	return "using: " + strings.Join(parts, " -> ")
}
