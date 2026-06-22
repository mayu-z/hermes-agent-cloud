package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Phosphor green palette — terminal.shop vibe
	ColorGreen      = lipgloss.Color("#465e4cff") // classic matrix green
	ColorGreenDim   = lipgloss.Color("#3c9151")
	ColorGreenFaint = lipgloss.Color("#003A0E")
	ColorAmber      = lipgloss.Color("#FFB000") // tool / highlight
	ColorRed        = lipgloss.Color("#FF2222")
	ColorWhite      = lipgloss.Color("#E8E8E8")
	ColorGray       = lipgloss.Color("#555555")
	ColorBg         = lipgloss.Color("#000000")
	ColorBgPanel    = lipgloss.Color("#030F03")

	// Base
	BaseStyle = lipgloss.NewStyle().
			Background(ColorBg).
			Foreground(ColorGreen)

	// Prompt prefix
	PromptStyle = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)

	// User message
	UserPrefixStyle = lipgloss.NewStyle().
			Foreground(ColorGreenDim).
			Bold(true)

	UserMsgStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	// Hermes message
	HermesPrefixStyle = lipgloss.NewStyle().
				Foreground(ColorGreen).
				Bold(true)

	HermesMsgStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	// Tool use
	ToolStyle = lipgloss.NewStyle().
			Foreground(ColorAmber).
			Italic(true)

	// Error
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	// Input line
	InputPrefixStyle = lipgloss.NewStyle().
				Foreground(ColorGreen).
				Bold(true)

	InputStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	// Status bar
	StatusStyle = lipgloss.NewStyle().
			Foreground(ColorGreenDim).
			Background(ColorBgPanel)

	// Border / divider
	DividerStyle = lipgloss.NewStyle().
			Foreground(ColorGreenFaint)

	// Wakeup / boot text
	BootStyle = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)

	BootDimStyle = lipgloss.NewStyle().
			Foreground(ColorGreenDim)

	// Sidebar
	SidebarStyle = lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorGreenDim).
			PaddingLeft(1)

	SidebarTitleStyle = lipgloss.NewStyle().
				Foreground(ColorGreen).
				Bold(true)

	SidebarLabelStyle = lipgloss.NewStyle().
				Foreground(ColorGreenDim)

	SidebarValueStyle = lipgloss.NewStyle().
				Foreground(ColorWhite)

	// Clock overlay
	ClockStyle = lipgloss.NewStyle().
			Foreground(ColorAmber).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAmber).
			Padding(0, 2)

	// Task priority
	PriorityHigh = lipgloss.NewStyle().Foreground(ColorRed).Bold(true)
	PriorityMed  = lipgloss.NewStyle().Foreground(ColorAmber)
	PriorityLow  = lipgloss.NewStyle().Foreground(ColorGreenDim)
)

var HermesASCII = `
 ██╗  ██╗███████╗██████╗ ███╗   ███╗███████╗███████╗
 ██║  ██║██╔════╝██╔══██╗████╗ ████║██╔════╝██╔════╝
 ███████║█████╗  ██████╔╝██╔████╔██║█████╗  ███████╗
 ██╔══██║██╔══╝  ██╔══██╗██║╚██╔╝██║██╔══╝  ╚════██║
 ██║  ██║███████╗██║  ██║██║ ╚═╝ ██║███████╗███████║
 ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝╚══════╝`

var BootLines = []string{
	"HERMES OS v2.0 — Personal AI Agent",
	"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
	"[OK] Loading neural core...",
	"[OK] Connecting to Cloudflare edge...",
	"[OK] Initialising memory (KV store)...",
	"[OK] Mounting tool registry...",
	"[OK] Calibrating personality matrix...",
	"[OK] Web search engine: READY",
	"[OK] All systems nominal.",
	"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
	"",
	"  HERMES IS ONLINE. Type your message below.",
	"  Press Ctrl+K for help.",
	"",
}
