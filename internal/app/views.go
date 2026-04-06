package app

import (
	"fmt"
	"math"
	"strings"
	"time"

	"vapt/internal/ui"
)

func spinnerRow(width, frame, offset int) string {
	spinChars := []rune{'◐', '◓', '◑', '◒'}
	amplitude := 1.0
	bandWidth := 4
	pad := int(amplitude) + 1
	outHeight := 1 + 2*pad

	row := make([]rune, width)
	for i := range row {
		row[i] = ' '
	}
	for col := 0; col < width; col += 3 {
		row[col] = spinChars[(frame+col/3+offset)%len(spinChars)]
	}

	padded := make([]rune, width)
	copy(padded, row)

	lines := make([]string, outHeight)
	for r := 0; r < outHeight; r++ {
		var b strings.Builder
		for col := 0; col < width; col++ {
			band := float64(col / bandWidth)
			phase := band*0.5 + float64(frame)*0.5
			off := int(math.Round(amplitude * math.Sin(phase)))
			srcRow := r - pad - off
			if srcRow == 0 {
				b.WriteRune(padded[col])
			} else {
				b.WriteRune(' ')
			}
		}
		lines[r] = b.String()
	}

	var s strings.Builder
	for _, l := range lines {
		s.WriteString("  " + l + "\n")
	}
	return s.String()
}

func (m Model) logo() string {
	f := m.animFrame
	width := 56

	var s strings.Builder
	s.WriteString(spinnerRow(width, f, 0))
	s.WriteString("\n")
	s.WriteString("      ██╗   ██╗  █████╗ ██████╗ ████████╗\n")
	s.WriteString("      ██║   ██║ ██╔══██╗██╔══██║╚══██╔══╝\n")
	s.WriteString("       ╚██╗ ██╔╝███████║██████╔╝   ██║\n")
	s.WriteString("        ╚████╔╝ ██╔══██║██╔═══╝    ██║\n")
	s.WriteString("         ╚═══╝  ██║  ██║██║       ██║\n")
	s.WriteString("\n")
	s.WriteString(spinnerRow(width, f, 2))
	s.WriteString("    ─ ── ─── >>> ⌨  t y p e  f a s t e r  >>> ─── ── ─")

	return s.String()
}

func (m Model) View() string {
	var s strings.Builder
	s.WriteString("\n")

	switch m.state {
	case stateMenu:
		s.WriteString(m.viewMenu())
	case stateTyping:
		s.WriteString(m.viewTyping())
	case stateResults:
		s.WriteString(m.viewResults())
	case stateLogin:
		s.WriteString(m.viewLogin())
	case stateRegister:
		s.WriteString(m.viewRegister())
	case stateLeaderboard:
		s.WriteString(m.viewLeaderboard())
	case stateTenant:
		s.WriteString(m.viewTenant())
	case stateProfile:
		s.WriteString(m.viewProfile())
	case stateMyResults:
		s.WriteString(m.viewMyResults())
	case stateTenantList:
		s.WriteString(m.viewTenantList())
	case stateJoinRequests:
		s.WriteString(m.viewJoinRequests())
	case stateInviteCodes:
		s.WriteString(m.viewInviteCodes())
	}

	if m.loading {
		frame := spinnerFrames[m.spinnerIdx%len(spinnerFrames)]
		s.WriteString("\n  " + ui.Styled(frame+" "+m.loadingMsg, ui.CurrentTheme.Accent, ui.Italic) + "\n")
	}

	s.WriteString("\n")
	return s.String()
}

func (m Model) viewMenu() string {
	b := ui.NewBox(0)

	subtitle := ui.Styled("terminal typing speed test", ui.CurrentTheme.TextDim, ui.Italic)
	b.Centered(subtitle)
	b.Sep()

	if !m.offline && m.user != nil {
		tenantInfo := ""
		if len(m.user.Tenants) > 0 {
			tenantInfo = " (" + m.activeTenantName() + ")"
		}
		b.Line("  " + ui.Styled("▸", ui.CurrentTheme.Success) + " logged in as: " +
			ui.Styled(m.user.Username+tenantInfo, ui.CurrentTheme.Accent))
	} else if m.offline {
		b.Line("  " + ui.Styled("▸ offline mode", ui.CurrentTheme.Warning))
	}
	b.Blank()

	type menuItem struct {
		key  string
		desc string
	}
	var items []menuItem

	items = append(items, menuItem{"0", "Start Test"})
	if !m.offline {
		if m.user != nil && len(m.user.Tenants) > 0 {
			items = append(items, menuItem{"1", "My Results"})
			items = append(items, menuItem{"2", "Leaderboard"})
			items = append(items, menuItem{"3", "Tenant"})
		}
	}
	n := len(items)
	items = append(items, menuItem{fmt.Sprintf("%d", n), fmt.Sprintf("Duration: %s", ui.Styled(fmt.Sprintf("%ds", m.duration), ui.CurrentTheme.Warning, ui.Bold))})
	if !m.offline {
		n = len(items)
		items = append(items, menuItem{fmt.Sprintf("%d", n), "Tenants"})
		if m.user != nil && len(m.user.Tenants) > 1 {
			n = len(items)
			items = append(items, menuItem{fmt.Sprintf("%d", n), "Switch Tenant"})
		}
		n = len(items)
		items = append(items, menuItem{fmt.Sprintf("%d", n), "Profile"})
		n = len(items)
		items = append(items, menuItem{fmt.Sprintf("%d", n), "Logout"})
	}
	n = len(items)
	items = append(items, menuItem{fmt.Sprintf("%d", n), "Quit"})

	for _, item := range items {
		b.Line("  " + ui.Styled("["+item.key+"]", ui.CurrentTheme.Accent, ui.Bold) + " " + item.desc)
	}
	b.Blank()

	if m.best.Score > 0 {
		b.Sep()
		b.Line("  " + ui.Styled("★", ui.CurrentTheme.Star) + " Best: " +
			ui.SparkScore(m.best.Score) + " SCR  " +
			ui.SparkWPM(m.best.WPM) + " WPM  " +
			ui.SparkAccuracy(m.best.Accuracy) + "  " +
			ui.Styled(fmt.Sprintf("(%ds)", m.best.Duration), ui.CurrentTheme.TextDim))
	}

	return ui.Styled(m.logo(), ui.CurrentTheme.Accent) + "\n" + b.Render()
}

func (m Model) viewTyping() string {
	var s strings.Builder
	s.WriteString(m.viewTimerBar())
	s.WriteString(m.viewSpeedAnim())
	s.WriteString(m.viewTextWindow())
	s.WriteString(m.viewLiveStats())
	return s.String()
}

func (m Model) viewTimerBar() string {
	remaining := m.timeLeft
	if remaining < 0 {
		remaining = 0
	}
	fraction := remaining / float64(m.duration)

	timeLabel := ui.Styled(fmt.Sprintf(" %02d:%02d ", int(remaining)/60, int(remaining)%60), ui.CurrentTheme.Text, ui.Bold)
	return "  " + ui.ProgressBar(fraction, 36) + timeLabel + "\n" +
		"  " + ui.Separator(44) + "\n\n"
}

func (m Model) viewSpeedAnim() string {
	liveWPM := 0.0
	if m.started {
		elapsed := time.Since(m.startTime).Minutes()
		if elapsed > 0 {
			liveWPM = (float64(len(m.typed)) / 5.0) / elapsed
		}
	}

	level := int(liveWPM / 16)
	if level > 5 {
		level = 5
	}

	width := 44
	f := m.animFrame
	var s strings.Builder
	s.WriteString("  ")

	if !m.started || level == 0 {
		for i := 0; i < width; i++ {
			if (i+f)%6 == 0 {
				s.WriteString(ui.Styled("·", ui.CurrentTheme.TextDim))
			} else {
				s.WriteString(" ")
			}
		}
	} else {
		tiers := []struct {
			chars  []rune
			density int
			speed   int
		}{
			{[]rune{'·'}, 8, 1},
			{[]rune{'·', '›'}, 6, 1},
			{[]rune{'›', '»', '─'}, 4, 2},
			{[]rune{'»', '─', '═'}, 3, 2},
			{[]rune{'═', '»', '▸', '─'}, 2, 3},
		}
		tier := tiers[level-1]
		for i := 0; i < width; i++ {
			pos := (i + f*tier.speed) % width
			if pos%tier.density == 0 {
				ch := tier.chars[pos%len(tier.chars)]
				color := ui.CurrentTheme.TextDim
				if level >= 3 {
					color = ui.CurrentTheme.Accent
				}
				if level >= 5 {
					color = ui.CurrentTheme.Warning
				}
				s.WriteString(ui.Styled(string(ch), color))
			} else {
				s.WriteString(" ")
			}
		}
	}
	s.WriteString("\n")
	return s.String()
}

func (m Model) viewTextWindow() string {
	var s strings.Builder

	windowStart := 0
	if len(m.typed) > 40 {
		windowStart = len(m.typed) - 40
	}
	windowEnd := windowStart + 80
	if windowEnd > len(m.target) {
		windowEnd = len(m.target)
	}

	s.WriteString("  ")
	for i := windowStart; i < windowEnd; i++ {
		ch := m.target[i]
		if i < len(m.typed) {
			if ch == m.typed[i] {
				s.WriteString(ui.Styled(string(ch), ui.CurrentTheme.Correct))
			} else {
				s.WriteString(ui.Styled(string(m.typed[i]), ui.CurrentTheme.Error, ui.CurrentTheme.ErrorBg, ui.CurrentTheme.Text))
			}
		} else if i == len(m.typed) {
			s.WriteString(ui.Styled(string(ch), ui.CurrentTheme.Cursor, ui.Underline, ui.Bold))
		} else {
			s.WriteString(ui.Styled(string(ch), ui.CurrentTheme.TextDim, ui.Dim))
		}
	}
	s.WriteString("\n\n")
	s.WriteString("  " + ui.Separator(44) + "\n")
	return s.String()
}

func (m Model) viewLiveStats() string {
	liveWPM := 0.0
	liveAcc := 100.0

	if m.started {
		elapsed := time.Since(m.startTime).Minutes()
		if elapsed > 0 {
			liveWPM = (float64(len(m.typed)) / 5.0) / elapsed
		}
		correct := 0
		for i := 0; i < len(m.typed) && i < len(m.target); i++ {
			if m.typed[i] == m.target[i] {
				correct++
			}
		}
		if len(m.typed) > 0 {
			liveAcc = float64(correct) / float64(len(m.typed)) * 100
		}
	}

	hint := ui.Styled("esc", ui.CurrentTheme.TextDim, ui.Dim) + ui.Styled(" cancel", ui.CurrentTheme.TextDim, ui.Dim)
	if !m.started {
		hint = ui.Styled("start typing...", ui.CurrentTheme.Accent, ui.Italic)
	}

	return "  " +
		ui.Styled("WPM ", ui.CurrentTheme.TextDim) + ui.SparkWPM(liveWPM) +
		ui.Styled("  │  ", ui.CurrentTheme.TextDim) +
		ui.Styled("ACC ", ui.CurrentTheme.TextDim) + ui.SparkAccuracy(liveAcc) +
		ui.Styled("  │  ", ui.CurrentTheme.TextDim) +
		hint + "\n"
}

func (m Model) viewResults() string {
	b := ui.NewBox(0)

	b.Centered(ui.Styled("Results", ui.CurrentTheme.Accent, ui.Bold))
	b.Sep()
	b.Blank()

	b.Line(ui.Styled("  WPM", ui.CurrentTheme.TextDim) + "  " + ui.SparkWPM(m.wpm))
	b.Line(ui.Styled("  ACC", ui.CurrentTheme.TextDim) + "  " + ui.SparkAccuracy(m.accuracy))
	b.Line(ui.Styled("  CON", ui.CurrentTheme.TextDim) + "  " + ui.SparkConsistency(m.consistency))
	b.Line(ui.Styled("  SCR", ui.CurrentTheme.TextDim) + "  " + ui.SparkScore(m.score))
	b.Line(ui.Styled("  DUR", ui.CurrentTheme.TextDim) + "  " + ui.Styled(fmt.Sprintf("%ds", m.duration), ui.CurrentTheme.Text))
	b.Line(ui.Styled("  CHR", ui.CurrentTheme.TextDim) + "  " + ui.Styled(fmt.Sprintf("%d", len(m.typed)), ui.CurrentTheme.Text))
	b.Blank()

	if m.score >= m.best.Score && m.score > 0 {
		b.Line(ui.Styled("  ★ New personal best!", ui.CurrentTheme.Star, ui.Bold))
	} else if m.best.Score > 0 {
		b.Line(ui.Styled(fmt.Sprintf("  Best: %.0f SCR (%.0f WPM)", m.best.Score, m.best.WPM), ui.CurrentTheme.TextDim))
	}
	if m.submitResp != nil {
		b.Line(ui.Styled("  ✦ Result submitted to server!", ui.CurrentTheme.Success))
	} else if m.submitErr != "" {
		b.Line(ui.Styled("  ✗ "+m.submitErr, ui.CurrentTheme.Warning))
	}
	b.Blank()

	rating := m.score / 120.0
	if rating > 1 {
		rating = 1
	}
	b.Line("  " + ui.ProgressBar(rating, 30) + ui.Styled(fmt.Sprintf(" %.0f", m.score), ui.CurrentTheme.Text, ui.Bold))
	b.Blank()

	b.Line(ui.Styled("  r", ui.CurrentTheme.Accent) + ui.Styled(" retake  ", ui.CurrentTheme.TextDim) +
		ui.Styled("enter", ui.CurrentTheme.Accent) + ui.Styled(" menu  ", ui.CurrentTheme.TextDim) +
		ui.Styled("q", ui.CurrentTheme.Accent) + ui.Styled(" quit", ui.CurrentTheme.TextDim))

	return b.Render()
}

func (m Model) viewLogin() string {
	b := ui.NewBox(0)

	b.Centered(ui.Styled("Login to VAPT", ui.CurrentTheme.Accent, ui.Bold))
	b.Sep()
	b.Blank()

	userCursor := " "
	if m.loginField == 0 {
		userCursor = ui.Styled("▸", ui.CurrentTheme.Accent)
	}
	userLabel := " " + userCursor + " Email: " + ui.Styled(m.loginEmail, ui.CurrentTheme.Text)
	if m.loginField == 0 {
		userLabel += ui.Styled("_", ui.CurrentTheme.Accent, ui.Blink)
	}
	b.Line(userLabel)

	passCursor := " "
	if m.loginField == 1 {
		passCursor = ui.Styled("▸", ui.CurrentTheme.Accent)
	}
	maskedPass := ui.Repeat("•", len(m.loginPass))
	passLabel := " " + passCursor + " Password: " + ui.Styled(maskedPass, ui.CurrentTheme.Text)
	if m.loginField == 1 {
		passLabel += ui.Styled("_", ui.CurrentTheme.Accent, ui.Blink)
	}
	b.Line(passLabel)
	b.Blank()

	if m.loginErr != "" {
		b.Line("  " + ui.Styled(m.loginErr, ui.CurrentTheme.Error))
		b.Blank()
	}

	b.Line(" " + ui.Styled("[enter]", ui.CurrentTheme.Accent) + " Login  " +
		ui.Styled("[tab]", ui.CurrentTheme.Accent) + " Register  " +
		ui.Styled("[esc]", ui.CurrentTheme.Accent) + " Offline")
	b.Blank()

	return ui.Styled(m.logo(), ui.CurrentTheme.Accent) + "\n" + b.Render()
}

func (m Model) viewRegister() string {
	b := ui.NewBox(0)

	b.Centered(ui.Styled("Register for VAPT", ui.CurrentTheme.Accent, ui.Bold))
	b.Sep()
	b.Blank()

	fields := []struct {
		label string
		value string
		mask  bool
	}{
		{"Email", m.regEmail, false},
		{"Username", m.regUsername, false},
		{"Password", m.regPass, true},
		{"Invite Code", m.regInviteCode, false},
	}

	for i, f := range fields {
		cursor := " "
		if m.regField == i {
			cursor = ui.Styled("▸", ui.CurrentTheme.Accent)
		}
		display := f.value
		if f.mask {
			display = ui.Repeat("•", len(f.value))
		}
		line := " " + cursor + " " + f.label + ": " + ui.Styled(display, ui.CurrentTheme.Text)
		if m.regField == i {
			line += ui.Styled("_", ui.CurrentTheme.Accent, ui.Blink)
		}
		b.Line(line)
	}

	b.Blank()

	if m.regErr != "" {
		b.Line("  " + ui.Styled(m.regErr, ui.CurrentTheme.Error))
		b.Blank()
	}

	b.Line(" " + ui.Styled("[enter]", ui.CurrentTheme.Accent) + " Register  " +
		ui.Styled("[esc]", ui.CurrentTheme.Accent) + " Back")
	b.Blank()

	return ui.Styled(m.logo(), ui.CurrentTheme.Accent) + "\n" + b.Render()
}

func (m Model) viewLeaderboard() string {
	b := ui.NewBox(0)

	metricLabel := "Score"
	if m.leaderboard != nil {
		switch m.leaderboard.Metric {
		case "wpm":
			metricLabel = "WPM"
		case "accuracy":
			metricLabel = "Accuracy"
		case "consistency":
			metricLabel = "Consistency"
		}
	}

	durLabel := "all"
	if m.lbDuration != "all" {
		durLabel = m.lbDuration + "s"
	}
	periodLabel := m.lbPeriod
	switch m.lbPeriod {
	case "all_time":
		periodLabel = "all time"
	}

	b.Centered(ui.Styled("Leaderboard ("+metricLabel+")", ui.CurrentTheme.Accent, ui.Bold))
	if tn := m.activeTenantName(); tn != "" {
		b.Line(ui.Styled("  tenant: "+tn, ui.CurrentTheme.TextDim, ui.Italic))
	}
	b.Line(ui.Styled("  "+durLabel+" · "+periodLabel, ui.CurrentTheme.TextDim, ui.Italic))
	b.Sep()

	if m.leaderboard != nil && len(m.leaderboard.Entries) > 0 {
		for _, e := range m.leaderboard.Entries {
			rank := ui.Styled(fmt.Sprintf("  #%-3d", e.Rank), ui.CurrentTheme.Info)
			name := ui.Styled(fmt.Sprintf("%-12s", e.User.Username), ui.CurrentTheme.Text)
			score := e.WPM * (e.Accuracy / 100) * (e.Accuracy / 100) * (e.Accuracy / 100)
			stats := ui.SparkScore(score) + " SCR  " + ui.SparkWPM(e.WPM) + " WPM  " + ui.SparkAccuracy(e.Accuracy)
			b.Line(rank + " " + name + " " + stats)
		}
	} else if m.leaderboard != nil {
		b.Line(ui.Styled("  No entries yet", ui.CurrentTheme.TextDim, ui.Italic))
	}

	b.Blank()
	b.Line(" " + ui.Styled("[s]", ui.CurrentTheme.Accent) + " Score  " +
		ui.Styled("[w]", ui.CurrentTheme.Accent) + " WPM  " +
		ui.Styled("[a]", ui.CurrentTheme.Accent) + " Acc  " +
		ui.Styled("[c]", ui.CurrentTheme.Accent) + " Con  " +
		ui.Styled("[d]", ui.CurrentTheme.Accent) + " Dur  " +
		ui.Styled("[p]", ui.CurrentTheme.Accent) + " Period  " +
		ui.Styled("[esc]", ui.CurrentTheme.Accent) + " Back")

	return b.Render()
}

func (m Model) viewTenant() string {
	b := ui.NewBox(0)

	b.Centered(ui.Styled("Tenant", ui.CurrentTheme.Accent, ui.Bold))
	b.Sep()

	if m.tenant != nil {
		b.Line("  Name: " + ui.Styled(m.tenant.Name, ui.CurrentTheme.Text, ui.Bold))
		b.Line("  Members: " + ui.Styled(fmt.Sprintf("%d", m.tenant.MemberCount), ui.CurrentTheme.Info))
		b.Blank()

		if len(m.tenant.Members) > 0 {
			b.Sep()
			for _, mem := range m.tenant.Members {
				name := ui.Styled(fmt.Sprintf("  %-14s", mem.Username), ui.CurrentTheme.Text)
				role := ui.Styled(mem.Role, ui.CurrentTheme.TextDim)
				b.Line(name + " " + role)
			}
		}
	} else {
		b.Line(ui.Styled("  Loading...", ui.CurrentTheme.TextDim, ui.Italic))
	}

	b.Blank()
	hint := " " + ui.Styled("[esc]", ui.CurrentTheme.Accent) + " Back"
	if m.activeTenantRole() == "owner" || m.activeTenantRole() == "admin" {
		hint = " " + ui.Styled("[j]", ui.CurrentTheme.Accent) + " Join Requests  " + ui.Styled("[esc]", ui.CurrentTheme.Accent) + " Back"
	}
	b.Line(hint)

	return b.Render()
}

func (m Model) viewProfile() string {
	b := ui.NewBox(0)

	b.Centered(ui.Styled("Profile", ui.CurrentTheme.Accent, ui.Bold))
	b.Sep()

	if m.user != nil {
		b.Line("  Email: " + ui.Styled(m.user.Email, ui.CurrentTheme.Text))
		b.Line("  Role: " + ui.Styled(m.user.Role, ui.CurrentTheme.Info))
		if m.user.CreatedAt != "" && len(m.user.CreatedAt) >= 10 {
			b.Line("  Since: " + ui.Styled(m.user.CreatedAt[:10], ui.CurrentTheme.TextDim))
		}
		if len(m.user.Tenants) > 0 {
			b.Blank()
			b.Line(ui.Styled("  Tenants:", ui.CurrentTheme.TextDim))
			for _, t := range m.user.Tenants {
				b.Line("    " + ui.Styled(t.Name, ui.CurrentTheme.Text) + " " + ui.Styled("("+t.Role+")", ui.CurrentTheme.TextDim))
			}
		}
	}

	b.Sep()

	uCursor := " "
	if m.profileField == 0 {
		uCursor = ui.Styled("▸", ui.CurrentTheme.Accent)
	}
	uLine := " " + uCursor + " Username: " + ui.Styled(m.profileUsername, ui.CurrentTheme.Text)
	if m.profileField == 0 {
		uLine += ui.Styled("_", ui.CurrentTheme.Accent, ui.Blink)
	}
	b.Line(uLine)

	pCursor := " "
	if m.profileField == 1 {
		pCursor = ui.Styled("▸", ui.CurrentTheme.Accent)
	}
	maskedPass := ui.Repeat("•", len(m.profilePassword))
	pLine := " " + pCursor + " New Password: " + ui.Styled(maskedPass, ui.CurrentTheme.Text)
	if m.profileField == 1 {
		pLine += ui.Styled("_", ui.CurrentTheme.Accent, ui.Blink)
	}
	b.Line(pLine)
	b.Blank()

	if m.profileMsg != "" {
		b.Line("  " + ui.Styled(m.profileMsg, ui.CurrentTheme.Success))
	}
	if m.profileErr != "" {
		b.Line("  " + ui.Styled(m.profileErr, ui.CurrentTheme.Error))
	}

	hint := " " + ui.Styled("[enter]", ui.CurrentTheme.Accent) + " Save  "
	if m.user != nil && m.user.Role == "admin" {
		hint += ui.Styled("[i]", ui.CurrentTheme.Accent) + " Invites  "
	}
	hint += ui.Styled("[esc]", ui.CurrentTheme.Accent) + " Back"
	b.Line(hint)

	return b.Render()
}

func (m Model) viewMyResults() string {
	b := ui.NewBox(0)

	b.Centered(ui.Styled("My Results", ui.CurrentTheme.Accent, ui.Bold))

	durLabel := "all"
	if m.myResultsFilter != "all" {
		durLabel = m.myResultsFilter + "s"
	}
	sortLabel := m.myResultsSort
	b.Line(ui.Styled("  "+durLabel+" · "+sortLabel+" · "+m.myResultsOrder, ui.CurrentTheme.TextDim, ui.Italic))
	b.Sep()

	if m.myResults != nil {
		if len(m.myResults.Data) > 0 {
			bestWPM := 0.0
			var sumWPM, sumAcc float64
			count := len(m.myResults.Data)
			last10 := count
			if last10 > 10 {
				last10 = 10
			}
			for i, r := range m.myResults.Data {
				if r.WPM > bestWPM {
					bestWPM = r.WPM
				}
				if i < last10 {
					sumWPM += r.WPM
					sumAcc += r.Accuracy
				}
			}
			avgWPM := sumWPM / float64(last10)
			avgAcc := sumAcc / float64(last10)

			b.Line("  Best: " + ui.SparkWPM(bestWPM) + " WPM  Avg: " + ui.SparkWPM(avgWPM) + " WPM")
			b.Line("  Acc: " + ui.SparkAccuracy(avgAcc) + "  Tests: " + ui.Styled(fmt.Sprintf("%d", m.myResults.Meta.Total), ui.CurrentTheme.Info))
			b.Sep()
		}

		if len(m.myResults.Data) > 0 {
			for _, r := range m.myResults.Data {
				date := r.PlayedAt
				if len(date) >= 10 {
					date = date[:10]
				}
				b.Line("  " + ui.SparkWPM(r.WPM) + " WPM " +
					ui.SparkAccuracy(r.Accuracy) + " " +
					ui.SparkConsistency(r.Consistency) + " " +
					ui.Styled(fmt.Sprintf("%ds", r.TestDuration), ui.CurrentTheme.TextDim) + " " +
					ui.Styled(date, ui.CurrentTheme.TextDim))
			}
		} else {
			b.Line(ui.Styled("  No results yet", ui.CurrentTheme.TextDim, ui.Italic))
		}

		if m.myResults.Meta.Total > 20 {
			b.Line(ui.Styled(fmt.Sprintf("  Page %d/%d", m.myResultsPage, (m.myResults.Meta.Total+19)/20), ui.CurrentTheme.TextDim))
		}
	}

	b.Blank()
	b.Line(" " + ui.Styled("[d]", ui.CurrentTheme.Accent) + " Dur  " +
		ui.Styled("[s]", ui.CurrentTheme.Accent) + " Sort  " +
		ui.Styled("[o]", ui.CurrentTheme.Accent) + " Order  " +
		ui.Styled("[n/b]", ui.CurrentTheme.Accent) + " Page  " +
		ui.Styled("[esc]", ui.CurrentTheme.Accent) + " Back")

	return b.Render()
}

func (m Model) viewTenantList() string {
	b := ui.NewBox(0)

	b.Centered(ui.Styled("Tenants", ui.CurrentTheme.Accent, ui.Bold))
	b.Sep()

	if m.tenantListData != nil && len(m.tenantListData.Data) > 0 {
		for i, t := range m.tenantListData.Data {
			cursor := "  "
			if i == m.tenantListCursor {
				cursor = ui.Styled(" ▸", ui.CurrentTheme.Accent)
			}
			name := ui.Styled(fmt.Sprintf("%-18s", t.Name), ui.CurrentTheme.Text)
			members := ui.Styled(fmt.Sprintf("%d members", t.MemberCount), ui.CurrentTheme.TextDim)
			b.Line(cursor + " " + name + " " + members)
		}
		if m.tenantListData.Meta.Total > 20 {
			b.Line(ui.Styled(fmt.Sprintf("  Page %d/%d", m.tenantListPage, (m.tenantListData.Meta.Total+19)/20), ui.CurrentTheme.TextDim))
		}
	} else if m.tenantListData != nil {
		b.Line(ui.Styled("  No tenants found", ui.CurrentTheme.TextDim, ui.Italic))
	}

	if m.tenantListMsg != "" {
		b.Blank()
		b.Line("  " + ui.Styled(m.tenantListMsg, ui.CurrentTheme.Info))
	}

	if m.tenantListCreating {
		b.Sep()
		b.Line(" " + ui.Styled("New tenant name:", ui.CurrentTheme.Accent))
		b.Line("  " + ui.Styled(m.tenantListName, ui.CurrentTheme.Text) + ui.Styled("_", ui.CurrentTheme.Accent, ui.Blink))
		b.Line("  " + ui.Styled("[enter]", ui.CurrentTheme.Accent) + " Create  " + ui.Styled("[esc]", ui.CurrentTheme.Accent) + " Cancel")
	}

	b.Blank()
	b.Line(" " + ui.Styled("[enter]", ui.CurrentTheme.Accent) + " View  " +
		ui.Styled("[j]", ui.CurrentTheme.Accent) + " Join  " +
		ui.Styled("[c]", ui.CurrentTheme.Accent) + " Create  " +
		ui.Styled("[n/b]", ui.CurrentTheme.Accent) + " Page  " +
		ui.Styled("[esc]", ui.CurrentTheme.Accent) + " Back")

	return b.Render()
}

func (m Model) viewJoinRequests() string {
	b := ui.NewBox(0)

	b.Centered(ui.Styled("Join Requests", ui.CurrentTheme.Accent, ui.Bold))
	b.Sep()

	if m.joinRequests != nil && len(m.joinRequests.Data) > 0 {
		for i, jr := range m.joinRequests.Data {
			cursor := "  "
			if i == m.joinRequestsCursor {
				cursor = ui.Styled(" ▸", ui.CurrentTheme.Accent)
			}
			date := jr.CreatedAt
			if len(date) >= 10 {
				date = date[:10]
			}
			name := ui.Styled(fmt.Sprintf("%-16s", jr.User.Username), ui.CurrentTheme.Text)
			b.Line(cursor + " " + name + " " + ui.Styled(date, ui.CurrentTheme.TextDim))
		}
	} else if m.joinRequests != nil {
		b.Line(ui.Styled("  No pending requests", ui.CurrentTheme.TextDim, ui.Italic))
	}

	if m.joinRequestsMsg != "" {
		b.Blank()
		b.Line("  " + ui.Styled(m.joinRequestsMsg, ui.CurrentTheme.Info))
	}

	b.Blank()
	b.Line(" " + ui.Styled("[a]", ui.CurrentTheme.Accent) + " Approve  " +
		ui.Styled("[r]", ui.CurrentTheme.Accent) + " Reject  " +
		ui.Styled("[esc]", ui.CurrentTheme.Accent) + " Back")

	return b.Render()
}

func (m Model) viewInviteCodes() string {
	b := ui.NewBox(0)

	b.Centered(ui.Styled("Invite Codes", ui.CurrentTheme.Accent, ui.Bold))
	b.Sep()

	if m.inviteCodeCreated != "" {
		b.Line(ui.Styled("  New: "+m.inviteCodeCreated, ui.CurrentTheme.Success, ui.Bold))
		b.Sep()
	}

	if m.inviteCodes != nil && len(m.inviteCodes.Data) > 0 {
		for i, ic := range m.inviteCodes.Data {
			cursor := "  "
			if i == m.inviteCodesCursor {
				cursor = ui.Styled(" ▸", ui.CurrentTheme.Accent)
			}
			code := ic.Code
			if len(code) > 8 {
				code = code[:8] + "..."
			}
			statusColor := ui.CurrentTheme.Success
			if ic.Status != "active" {
				statusColor = ui.CurrentTheme.TextDim
			}
			by := ""
			if ic.CreatedBy.Username != "" {
				by = ic.CreatedBy.Username
			}
			b.Line(cursor + " " + ui.Styled(code, ui.CurrentTheme.Text) + " " +
				ui.Styled(fmt.Sprintf("%-8s", by), ui.CurrentTheme.TextDim) + " " +
				ui.Styled(ic.Status, statusColor))
		}
	} else if m.inviteCodes != nil {
		b.Line(ui.Styled("  No invite codes", ui.CurrentTheme.TextDim, ui.Italic))
	}

	if m.inviteCodesMsg != "" {
		b.Blank()
		b.Line("  " + ui.Styled(m.inviteCodesMsg, ui.CurrentTheme.Info))
	}

	b.Blank()
	b.Line(" " + ui.Styled("[c]", ui.CurrentTheme.Accent) + " Create  " +
		ui.Styled("[x]", ui.CurrentTheme.Accent) + " Invalidate  " +
		ui.Styled("[esc]", ui.CurrentTheme.Accent) + " Back")

	return b.Render()
}
