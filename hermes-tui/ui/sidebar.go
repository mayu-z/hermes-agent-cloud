package ui

import (
	"fmt"
	"strings"

	"hermes-tui/client"
)

func RenderSidebar(profile *client.Profile, taskCount int, sessionID string, connected bool, height int) string {
	var sb strings.Builder

	sb.WriteString(SidebarTitleStyle.Render("[ HERMES ]"))
	sb.WriteString("\n")
	sb.WriteString(DividerStyle.Render(strings.Repeat("─", 22)))
	sb.WriteString("\n\n")

	// Status
	if connected {
		sb.WriteString(PromptStyle.Render("● ONLINE"))
	} else {
		sb.WriteString(ErrorStyle.Render("● OFFLINE"))
	}
	sb.WriteString("\n\n")

	// User
	sb.WriteString(SidebarLabelStyle.Render("USER  : "))
	name := "unknown"
	if profile != nil && profile.Name != "" {
		name = profile.Name
	}
	sb.WriteString(SidebarValueStyle.Render(name))
	sb.WriteString("\n")

	// Session
	sid := sessionID
	if len(sid) > 14 {
		sid = sid[:14] + "…"
	}
	sb.WriteString(SidebarLabelStyle.Render("SESS  : "))
	sb.WriteString(SidebarValueStyle.Render(sid))
	sb.WriteString("\n")

	// Tasks
	sb.WriteString(SidebarLabelStyle.Render("TASKS : "))
	if taskCount > 0 {
		sb.WriteString(PriorityMed.Render(fmt.Sprintf("%d active", taskCount)))
	} else {
		sb.WriteString(SidebarValueStyle.Render("0 active"))
	}
	sb.WriteString("\n\n")

	// Interests
	if profile != nil && len(profile.Interests) > 0 {
		sb.WriteString(DividerStyle.Render(strings.Repeat("─", 22)))
		sb.WriteString("\n")
		sb.WriteString(SidebarLabelStyle.Render("KNOWS:\n"))
		for i, v := range profile.Interests {
			if i >= 4 {
				sb.WriteString(SidebarValueStyle.Render(fmt.Sprintf("  +%d more\n", len(profile.Interests)-4)))
				break
			}
			sb.WriteString(PromptStyle.Render("  > "))
			sb.WriteString(SidebarValueStyle.Render(v + "\n"))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(DividerStyle.Render(strings.Repeat("─", 22)))
	sb.WriteString("\n")
	sb.WriteString(SidebarLabelStyle.Render("KEYS:\n"))
	keys := []string{
		"Enter   send",
		"Ctrl+K  help",
		"Ctrl+L  clear",
		"Ctrl+T  tasks",
		"Ctrl+H  files",
		"Ctrl+P  clock",
		"Ctrl+C  quit",
		"↑↓      scroll",
	}
	for _, k := range keys {
		sb.WriteString(SidebarValueStyle.Render("  " + k + "\n"))
	}

	return SidebarStyle.Render(sb.String())
}