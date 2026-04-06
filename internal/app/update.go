package app

import (
	"math"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"vapt/internal/api"
	"vapt/internal/auth"
	"vapt/internal/storage"
	"vapt/internal/words"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) handle401(err error) (Model, bool) {
	if api.Is401(err) {
		auth.ClearToken()
		m.token = ""
		m.user = nil
		m.offline = true
		m.client.SetToken("")
		m.state = stateLogin
		m.loginErr = "Session expired, please login again"
		m.loginEmail = ""
		m.loginPass = ""
		m.loginField = 0
		m.loading = false
		return m, true
	}
	return m, false
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func spinnerTickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

type spinnerTickMsg struct{}

type animTickMsg struct{}

func animTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return animTickMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(spinnerTickMsg); ok {
		if m.loading {
			m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
			return m, spinnerTickCmd()
		}
		return m, nil
	}

	if _, ok := msg.(animTickMsg); ok {
		m.animFrame++
		return m, animTickCmd()
	}

	if m.loading {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case validateTokenMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			if m.token != "" {
				m.offline = false
				m.state = stateMenu
				return m, nil
			}
			m.offline = true
			m.state = stateMenu
			return m, nil
		}
		m.user = msg.user
		m.offline = false
		m.resolveActiveTenant()
		m.state = stateMenu
		return m, nil
	case submitResultMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.submitErr = msg.err.Error()
			return m, nil
		}
		m.submitResp = msg.resp
		m.submitErr = ""
		return m, nil
	case logoutMsg:
		auth.ClearToken()
		m.token = ""
		m.user = nil
		m.offline = true
		m.client.SetToken("")
		m.state = stateLogin
		m.loginEmail = ""
		m.loginPass = ""
		m.loginErr = ""
		m.loginField = 0
		m.loading = false
		return m, nil
	case updateProfileMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.profileErr = msg.err.Error()
			return m, nil
		}
		m.user = msg.user
		m.profileMsg = "Profile updated!"
		m.profileErr = ""
		return m, nil
	case myResultsMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.state = stateMenu
			return m, nil
		}
		m.myResults = msg.data
		return m, nil
	case tenantListMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.state = stateMenu
			return m, nil
		}
		m.tenantListData = msg.data
		return m, nil
	case createTenantMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.tenantListMsg = msg.err.Error()
			return m, nil
		}
		m.tenantListMsg = "Tenant created!"
		m.loading = true
		m.loadingMsg = "Refreshing..."
		return m, tea.Batch(fetchTenantListCmd(m.client, "", m.tenantListPage, 20), spinnerTickCmd())
	case joinRequestsMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.state = stateTenant
			return m, nil
		}
		m.joinRequests = msg.data
		return m, nil
	case resolveJoinRequestMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.joinRequestsMsg = msg.err.Error()
			return m, nil
		}
		m.joinRequestsMsg = "Done!"
		tid := m.activeTenantID()
		if tid != "" {
			m.loading = true
			m.loadingMsg = "Refreshing..."
			return m, tea.Batch(fetchJoinRequestsCmd(m.client, tid, "pending", 1, 20), spinnerTickCmd())
		}
		return m, nil
	case sendJoinRequestMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.tenantListMsg = msg.err.Error()
			return m, nil
		}
		m.tenantListMsg = "Join request sent!"
		return m, nil
	case inviteCodesMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.state = stateProfile
			return m, nil
		}
		m.inviteCodes = msg.data
		return m, nil
	case createInviteCodeMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.inviteCodesMsg = msg.err.Error()
			return m, nil
		}
		if msg.code != nil {
			m.inviteCodeCreated = msg.code.Code
		}
		m.loading = true
		m.loadingMsg = "Refreshing..."
		return m, tea.Batch(fetchInviteCodesCmd(m.client, 1, 20), spinnerTickCmd())
	case invalidateInviteCodeMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.inviteCodesMsg = msg.err.Error()
			return m, nil
		}
		m.inviteCodesMsg = "Code invalidated!"
		m.loading = true
		m.loadingMsg = "Refreshing..."
		return m, tea.Batch(fetchInviteCodesCmd(m.client, 1, 20), spinnerTickCmd())
	}

	switch m.state {
	case stateMenu:
		return m.menuUpdate(msg)
	case stateTyping:
		return m.typingUpdate(msg)
	case stateResults:
		return m.resultsUpdate(msg)
	case stateLogin:
		return m.loginUpdate(msg)
	case stateRegister:
		return m.registerUpdate(msg)
	case stateLeaderboard:
		return m.leaderboardUpdate(msg)
	case stateTenant:
		return m.tenantUpdate(msg)
	case stateProfile:
		return m.profileUpdate(msg)
	case stateMyResults:
		return m.myResultsUpdate(msg)
	case stateTenantList:
		return m.tenantListUpdate(msg)
	case stateJoinRequests:
		return m.joinRequestsUpdate(msg)
	case stateInviteCodes:
		return m.inviteCodesUpdate(msg)
	}
	return m, nil
}

func (m Model) menuUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		idx := m.menuKeyIndex(key)
		if idx < 0 {
			return m, nil
		}
		action := m.menuActions()[idx]
		switch action {
		case "start":
			m.state = stateTyping
			m.target = words.GenerateText(500)
			m.typed = ""
			m.started = false
			m.timeLeft = float64(m.duration)
			m.submitResp = nil
			m.wordTimes = nil
			return m, nil
		case "my_results":
			m.state = stateMyResults
			m.myResultsPage = 1
			m.myResults = nil
			m.loading = true
			m.loadingMsg = "Loading results..."
			return m, tea.Batch(fetchMyResultsCmd(m.client, m.activeTenantID(), m.myResultsPage, 20, m.myResultsFilter, m.myResultsSort, m.myResultsOrder), spinnerTickCmd())
		case "leaderboard":
			m.state = stateLeaderboard
			m.loading = true
			m.loadingMsg = "Loading leaderboard..."
			return m, tea.Batch(fetchLeaderboardCmd(m.client, m.activeTenantID(), m.lbAPIMetric(), m.lbDuration, m.lbPeriod, 50), spinnerTickCmd())
		case "tenant":
			m.state = stateTenant
			m.loading = true
			m.loadingMsg = "Loading tenant..."
			return m, tea.Batch(fetchTenantCmd(m.client, m.activeTenantID()), spinnerTickCmd())
		case "duration":
			m.durationIdx = (m.durationIdx + 1) % len(durations)
			m.duration = durations[m.durationIdx]
			return m, nil
		case "tenants":
			m.state = stateTenantList
			m.tenantListData = nil
			m.tenantListCursor = 0
			m.tenantListPage = 1
			m.tenantListMsg = ""
			m.tenantListCreating = false
			m.tenantListName = ""
			m.loading = true
			m.loadingMsg = "Loading tenants..."
			return m, tea.Batch(fetchTenantListCmd(m.client, "", 1, 20), spinnerTickCmd())
		case "switch_tenant":
			m.activeTenantIdx = (m.activeTenantIdx + 1) % len(m.user.Tenants)
			auth.SaveActiveTenant(m.user.Tenants[m.activeTenantIdx].TenantID)
			return m, nil
		case "profile":
			m.state = stateProfile
			m.profileUsername = m.user.Username
			m.profilePassword = ""
			m.profileField = 0
			m.profileMsg = ""
			m.profileErr = ""
			return m, nil
		case "logout":
			m.loading = true
			m.loadingMsg = "Logging out..."
			return m, tea.Batch(logoutCmd(m.client), spinnerTickCmd())
		case "quit":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) finishTest() Model {
	elapsed := time.Since(m.startTime).Minutes()
	w := float64(len(m.typed)) / 5.0
	if elapsed > 0 {
		m.wpm = w / elapsed
	}
	correct := 0
	for i := 0; i < len(m.typed) && i < len(m.target); i++ {
		if m.typed[i] == m.target[i] {
			correct++
		}
	}
	if len(m.typed) > 0 {
		m.accuracy = float64(correct) / float64(len(m.typed)) * 100
	}

	if len(m.wordTimes) >= 2 {
		var sum float64
		for _, t := range m.wordTimes {
			sum += t
		}
		mean := sum / float64(len(m.wordTimes))
		var variance float64
		for _, t := range m.wordTimes {
			d := t - mean
			variance += d * d
		}
		stddev := math.Sqrt(variance / float64(len(m.wordTimes)))
		if mean > 0 {
			cv := stddev / mean
			m.consistency = math.Max(0, 100-cv*100)
		} else {
			m.consistency = 100
		}
	} else {
		m.consistency = 100
	}

	m.score = m.wpm * math.Pow(m.accuracy/100, 3)
	if m.score > m.best.Score {
		m.best = storage.BestResult{WPM: m.wpm, Accuracy: m.accuracy, Score: m.score, Duration: m.duration}
		storage.SaveBest(m.best)
	}
	m.state = stateResults
	return m
}

func (m Model) typingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if !m.started {
			return m, nil
		}
		m.timeLeft = float64(m.duration) - time.Since(m.startTime).Seconds()
		if m.timeLeft <= 0 {
			m.timeLeft = 0
			m = m.finishTest()
			if m.offline {
				m.submitErr = "offline — result not uploaded"
			} else if m.user == nil {
				m.submitErr = "not logged in — result not uploaded"
			} else if len(m.user.Tenants) == 0 {
				m.submitErr = "no tenant membership — result not uploaded"
			} else {
				req := api.SubmitResultReq{
					WPM:          m.wpm,
					Accuracy:     m.accuracy,
					Consistency:  m.consistency,
					TestDuration: m.duration,
				}
				m.submitErr = ""
				return m, submitResultCmd(m.client, m.activeTenantID(), req)
			}
			return m, nil
		}
		return m, tickCmd()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateMenu
			return m, nil
		case "backspace":
			if len(m.typed) > 0 {
				m.typed = m.typed[:len(m.typed)-1]
			}
			return m, nil
		default:
			key := msg.String()
			if len(key) == 1 {
				m.typed += key
			} else if key == "space" {
				m.typed += " "
			} else {
				return m, nil
			}
			startTick := false
			if !m.started {
				m.started = true
				m.startTime = time.Now()
				m.lastWordTime = m.startTime
				m.timeLeft = float64(m.duration)
				startTick = true
			}
			pos := len(m.typed) - 1
			if pos >= 0 && pos < len(m.target) && m.typed[pos] == ' ' && m.target[pos] == ' ' {
				now := time.Now()
				wordDur := now.Sub(m.lastWordTime).Seconds()
				if wordDur > 0 {
					m.wordTimes = append(m.wordTimes, wordDur)
				}
				m.lastWordTime = now
			}
			if len(m.typed) > len(m.target)-100 {
				m.target += " " + words.GenerateText(300)
			}
			if startTick {
				return m, tickCmd()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m Model) resultsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m.state = stateTyping
			m.target = words.GenerateText(500)
			m.typed = ""
			m.started = false
			m.timeLeft = float64(m.duration)
			m.wpm = 0
			m.accuracy = 0
			m.consistency = 0
			m.score = 0
			m.submitResp = nil
			m.wordTimes = nil
			return m, nil
		case "enter":
			m.state = stateMenu
			return m, nil
		}
	}
	return m, nil
}

func (m Model) loginUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loginResultMsg:
		m.loading = false
		if msg.err != nil {
			m.loginErr = msg.err.Error()
			return m, nil
		}
		m.token = msg.token
		m.user = msg.user
		m.offline = false
		m.client.SetToken(msg.token)
		auth.SaveToken(msg.token)
		m.state = stateMenu
		m.loginEmail = ""
		m.loginPass = ""
		m.loginErr = ""
		return m, validateTokenCmd(m.client)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.offline = true
			m.state = stateMenu
			return m, nil
		case "tab":
			m.state = stateRegister
			m.regEmail = m.loginEmail
			m.regUsername = ""
			m.regPass = m.loginPass
			m.regField = 0
			m.regErr = ""
			return m, nil
		case "enter":
			if m.loginEmail == "" || m.loginPass == "" {
				m.loginErr = "Email and password required"
				return m, nil
			}
			m.loading = true
			m.loadingMsg = "Logging in..."
			m.loginErr = ""
			return m, tea.Batch(loginCmd(m.client, m.loginEmail, m.loginPass), spinnerTickCmd())
		case "shift+tab", "up", "down":
			m.loginField = 1 - m.loginField
			return m, nil
		case "backspace":
			if m.loginField == 0 && len(m.loginEmail) > 0 {
				m.loginEmail = m.loginEmail[:len(m.loginEmail)-1]
			} else if m.loginField == 1 && len(m.loginPass) > 0 {
				m.loginPass = m.loginPass[:len(m.loginPass)-1]
			}
			return m, nil
		default:
			key := msg.String()
			ch := key
			if key == "space" {
				ch = " "
			}
			if len(ch) != 1 {
				return m, nil
			}
			if m.loginField == 0 {
				m.loginEmail = auth.SanitizeInput(m.loginEmail + ch)
			} else {
				m.loginPass = auth.SanitizeInput(m.loginPass + ch)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m Model) registerUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case registerResultMsg:
		m.loading = false
		if msg.err != nil {
			m.regErr = msg.err.Error()
			return m, nil
		}
		m.token = msg.token
		m.user = msg.user
		m.offline = false
		m.client.SetToken(msg.token)
		auth.SaveToken(msg.token)
		m.state = stateMenu
		m.regEmail = ""
		m.regUsername = ""
		m.regPass = ""
		m.regInviteCode = ""
		m.regErr = ""
		return m, validateTokenCmd(m.client)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateLogin
			m.loginEmail = m.regEmail
			m.loginPass = m.regPass
			m.loginField = 0
			m.loginErr = ""
			return m, nil
		case "enter":
			if m.regEmail == "" || m.regUsername == "" || m.regPass == "" {
				m.regErr = "Email, username, and password required"
				return m, nil
			}
			m.loading = true
			m.loadingMsg = "Registering..."
			m.regErr = ""
			return m, tea.Batch(registerCmd(m.client, m.regEmail, m.regUsername, m.regPass, m.regInviteCode), spinnerTickCmd())
		case "tab", "down":
			m.regField = (m.regField + 1) % 4
			return m, nil
		case "shift+tab", "up":
			m.regField = (m.regField + 3) % 4
			return m, nil
		case "backspace":
			switch m.regField {
			case 0:
				if len(m.regEmail) > 0 {
					m.regEmail = m.regEmail[:len(m.regEmail)-1]
				}
			case 1:
				if len(m.regUsername) > 0 {
					m.regUsername = m.regUsername[:len(m.regUsername)-1]
				}
			case 2:
				if len(m.regPass) > 0 {
					m.regPass = m.regPass[:len(m.regPass)-1]
				}
			case 3:
				if len(m.regInviteCode) > 0 {
					m.regInviteCode = m.regInviteCode[:len(m.regInviteCode)-1]
				}
			}
			return m, nil
		default:
			key := msg.String()
			ch := key
			if key == "space" {
				ch = " "
			}
			if len(ch) != 1 {
				return m, nil
			}
			switch m.regField {
			case 0:
				m.regEmail = auth.SanitizeInput(m.regEmail + ch)
			case 1:
				m.regUsername = auth.SanitizeInput(m.regUsername + ch)
			case 2:
				m.regPass = auth.SanitizeInput(m.regPass + ch)
			case 3:
				m.regInviteCode = auth.SanitizeInput(m.regInviteCode + ch)
			}
			return m, nil
		}
	}
	return m, nil
}

// lbAPIMetric translates client metric to API metric ("score" → "wpm")
func (m Model) lbAPIMetric() string {
	if m.lbMetric == "score" {
		return "wpm"
	}
	return m.lbMetric
}

func (m Model) leaderboardUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case leaderboardMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.state = stateMenu
			return m, nil
		}
		m.leaderboard = msg.data
		if m.lbMetric == "score" && m.leaderboard != nil {
			entries := m.leaderboard.Entries
			sort.Slice(entries, func(i, j int) bool {
				si := entries[i].WPM * math.Pow(entries[i].Accuracy/100, 3)
				sj := entries[j].WPM * math.Pow(entries[j].Accuracy/100, 3)
				return si > sj
			})
			for i := range entries {
				entries[i].Rank = i + 1
			}
			m.leaderboard.Entries = entries
			m.leaderboard.Metric = "score"
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateMenu
			m.leaderboard = nil
			return m, nil
		case "s", "w", "a", "c":
			if m.user == nil || len(m.user.Tenants) == 0 {
				return m, nil
			}
			switch msg.String() {
			case "s":
				m.lbMetric = "score"
			case "w":
				m.lbMetric = "wpm"
			case "a":
				m.lbMetric = "accuracy"
			case "c":
				m.lbMetric = "consistency"
			}
			m.loading = true
			m.loadingMsg = "Loading leaderboard..."
			return m, tea.Batch(fetchLeaderboardCmd(m.client, m.activeTenantID(), m.lbAPIMetric(), m.lbDuration, m.lbPeriod, 50), spinnerTickCmd())
		case "d":
			if m.user == nil || len(m.user.Tenants) == 0 {
				return m, nil
			}
			durs := []string{"all", "15", "30", "60", "120"}
			for i, d := range durs {
				if d == m.lbDuration {
					m.lbDuration = durs[(i+1)%len(durs)]
					break
				}
			}
			m.loading = true
			m.loadingMsg = "Loading leaderboard..."
			return m, tea.Batch(fetchLeaderboardCmd(m.client, m.activeTenantID(), m.lbAPIMetric(), m.lbDuration, m.lbPeriod, 50), spinnerTickCmd())
		case "p":
			if m.user == nil || len(m.user.Tenants) == 0 {
				return m, nil
			}
			periods := []string{"all_time", "monthly", "weekly", "daily"}
			for i, p := range periods {
				if p == m.lbPeriod {
					m.lbPeriod = periods[(i+1)%len(periods)]
					break
				}
			}
			m.loading = true
			m.loadingMsg = "Loading leaderboard..."
			return m, tea.Batch(fetchLeaderboardCmd(m.client, m.activeTenantID(), m.lbAPIMetric(), m.lbDuration, m.lbPeriod, 50), spinnerTickCmd())
		}
	}
	return m, nil
}

func (m Model) tenantUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tenantMsg:
		m.loading = false
		if msg.err != nil {
			if m2, handled := m.handle401(msg.err); handled {
				return m2, nil
			}
			m.state = stateMenu
			return m, nil
		}
		m.tenant = msg.data
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateMenu
			m.tenant = nil
			return m, nil
		case "j":
			role := m.activeTenantRole()
			if role == "owner" || role == "admin" {
				tid := m.activeTenantID()
				if tid != "" {
					m.state = stateJoinRequests
					m.joinRequests = nil
					m.joinRequestsCursor = 0
					m.joinRequestsMsg = ""
					m.loading = true
					m.loadingMsg = "Loading join requests..."
					return m, tea.Batch(fetchJoinRequestsCmd(m.client, tid, "pending", 1, 20), spinnerTickCmd())
				}
			}
			return m, nil
		}
	}
	return m, nil
}

func (m Model) profileUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateMenu
			return m, nil
		case "tab", "down":
			m.profileField = 1 - m.profileField
			return m, nil
		case "shift+tab", "up":
			m.profileField = 1 - m.profileField
			return m, nil
		case "i":
			if m.user != nil && m.user.Role == "admin" {
				m.state = stateInviteCodes
				m.inviteCodes = nil
				m.inviteCodesCursor = 0
				m.inviteCodeCreated = ""
				m.inviteCodesMsg = ""
				m.loading = true
				m.loadingMsg = "Loading invite codes..."
				return m, tea.Batch(fetchInviteCodesCmd(m.client, 1, 20), spinnerTickCmd())
			}
			return m, nil
		case "enter":
			req := api.UpdateProfileReq{}
			if m.profileUsername != "" && m.profileUsername != m.user.Username {
				req.Username = m.profileUsername
			}
			if m.profilePassword != "" {
				req.Password = m.profilePassword
			}
			if req.Username == "" && req.Password == "" {
				m.profileErr = "No changes to save"
				return m, nil
			}
			m.loading = true
			m.loadingMsg = "Saving..."
			m.profileMsg = ""
			m.profileErr = ""
			return m, tea.Batch(updateProfileCmd(m.client, req), spinnerTickCmd())
		case "backspace":
			if m.profileField == 0 && len(m.profileUsername) > 0 {
				m.profileUsername = m.profileUsername[:len(m.profileUsername)-1]
			} else if m.profileField == 1 && len(m.profilePassword) > 0 {
				m.profilePassword = m.profilePassword[:len(m.profilePassword)-1]
			}
			return m, nil
		default:
			key := msg.String()
			ch := key
			if key == "space" {
				ch = " "
			}
			if len(ch) != 1 {
				return m, nil
			}
			if m.profileField == 0 {
				m.profileUsername = auth.SanitizeInput(m.profileUsername + ch)
			} else {
				m.profilePassword = auth.SanitizeInput(m.profilePassword + ch)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m Model) myResultsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateMenu
			m.myResults = nil
			return m, nil
		case "d":
			filters := []string{"all", "15", "30", "60", "120"}
			for i, f := range filters {
				if f == m.myResultsFilter {
					m.myResultsFilter = filters[(i+1)%len(filters)]
					break
				}
			}
			m.myResultsPage = 1
			m.loading = true
			m.loadingMsg = "Loading results..."
			return m, tea.Batch(fetchMyResultsCmd(m.client, m.activeTenantID(), m.myResultsPage, 20, m.myResultsFilter, m.myResultsSort, m.myResultsOrder), spinnerTickCmd())
		case "s":
			sorts := []string{"played_at", "wpm", "accuracy", "consistency"}
			for i, s := range sorts {
				if s == m.myResultsSort {
					m.myResultsSort = sorts[(i+1)%len(sorts)]
					break
				}
			}
			m.myResultsPage = 1
			m.loading = true
			m.loadingMsg = "Loading results..."
			return m, tea.Batch(fetchMyResultsCmd(m.client, m.activeTenantID(), m.myResultsPage, 20, m.myResultsFilter, m.myResultsSort, m.myResultsOrder), spinnerTickCmd())
		case "o":
			if m.myResultsOrder == "desc" {
				m.myResultsOrder = "asc"
			} else {
				m.myResultsOrder = "desc"
			}
			m.myResultsPage = 1
			m.loading = true
			m.loadingMsg = "Loading results..."
			return m, tea.Batch(fetchMyResultsCmd(m.client, m.activeTenantID(), m.myResultsPage, 20, m.myResultsFilter, m.myResultsSort, m.myResultsOrder), spinnerTickCmd())
		case "n":
			if m.myResults != nil && m.myResultsPage*20 < m.myResults.Meta.Total {
				m.myResultsPage++
				m.loading = true
				m.loadingMsg = "Loading results..."
				return m, tea.Batch(fetchMyResultsCmd(m.client, m.activeTenantID(), m.myResultsPage, 20, m.myResultsFilter, m.myResultsSort, m.myResultsOrder), spinnerTickCmd())
			}
			return m, nil
		case "b":
			if m.myResultsPage > 1 {
				m.myResultsPage--
				m.loading = true
				m.loadingMsg = "Loading results..."
				return m, tea.Batch(fetchMyResultsCmd(m.client, m.activeTenantID(), m.myResultsPage, 20, m.myResultsFilter, m.myResultsSort, m.myResultsOrder), spinnerTickCmd())
			}
			return m, nil
		}
	}
	return m, nil
}

func (m Model) tenantListUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.tenantListCreating {
			switch msg.String() {
			case "esc":
				m.tenantListCreating = false
				m.tenantListName = ""
				return m, nil
			case "enter":
				name := m.tenantListName
				if name == "" {
					return m, nil
				}
				m.tenantListCreating = false
				m.tenantListName = ""
				m.loading = true
				m.loadingMsg = "Creating tenant..."
				m.tenantListMsg = ""
				return m, tea.Batch(createTenantCmd(m.client, name), spinnerTickCmd())
			case "backspace":
				if len(m.tenantListName) > 0 {
					m.tenantListName = m.tenantListName[:len(m.tenantListName)-1]
				}
				return m, nil
			default:
				if len(msg.String()) == 1 {
					m.tenantListName += msg.String()
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateMenu
			m.tenantListData = nil
			m.tenantListMsg = ""
			return m, nil
		case "up", "k":
			if m.tenantListCursor > 0 {
				m.tenantListCursor--
			}
			return m, nil
		case "down":
			if m.tenantListData != nil && m.tenantListCursor < len(m.tenantListData.Data)-1 {
				m.tenantListCursor++
			}
			return m, nil
		case "enter":
			if m.tenantListData != nil && len(m.tenantListData.Data) > 0 {
				selected := m.tenantListData.Data[m.tenantListCursor]
				m.state = stateTenant
				m.tenant = nil
				m.loading = true
				m.loadingMsg = "Loading tenant..."
				return m, tea.Batch(fetchTenantCmd(m.client, selected.ID), spinnerTickCmd())
			}
			return m, nil
		case "j":
			if m.tenantListData != nil && len(m.tenantListData.Data) > 0 {
				selected := m.tenantListData.Data[m.tenantListCursor]
				m.loading = true
				m.loadingMsg = "Sending join request..."
				m.tenantListMsg = ""
				return m, tea.Batch(sendJoinRequestCmd(m.client, selected.ID), spinnerTickCmd())
			}
			return m, nil
		case "n":
			if m.tenantListData != nil && m.tenantListPage*20 < m.tenantListData.Meta.Total {
				m.tenantListPage++
				m.tenantListCursor = 0
				m.loading = true
				m.loadingMsg = "Loading tenants..."
				return m, tea.Batch(fetchTenantListCmd(m.client, "", m.tenantListPage, 20), spinnerTickCmd())
			}
			return m, nil
		case "b":
			if m.tenantListPage > 1 {
				m.tenantListPage--
				m.tenantListCursor = 0
				m.loading = true
				m.loadingMsg = "Loading tenants..."
				return m, tea.Batch(fetchTenantListCmd(m.client, "", m.tenantListPage, 20), spinnerTickCmd())
			}
			return m, nil
		case "c":
			m.tenantListCreating = true
			m.tenantListName = ""
			return m, nil
		}
	}
	return m, nil
}

func (m Model) joinRequestsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateTenant
			m.joinRequests = nil
			m.joinRequestsMsg = ""
			return m, nil
		case "up", "k":
			if m.joinRequestsCursor > 0 {
				m.joinRequestsCursor--
			}
			return m, nil
		case "down":
			if m.joinRequests != nil && m.joinRequestsCursor < len(m.joinRequests.Data)-1 {
				m.joinRequestsCursor++
			}
			return m, nil
		case "a":
			if m.joinRequests != nil && len(m.joinRequests.Data) > 0 {
				req := m.joinRequests.Data[m.joinRequestsCursor]
				tid := m.activeTenantID()
				m.loading = true
				m.loadingMsg = "Approving..."
				m.joinRequestsMsg = ""
				return m, tea.Batch(resolveJoinRequestCmd(m.client, tid, req.ID, "approved"), spinnerTickCmd())
			}
			return m, nil
		case "r":
			if m.joinRequests != nil && len(m.joinRequests.Data) > 0 {
				req := m.joinRequests.Data[m.joinRequestsCursor]
				tid := m.activeTenantID()
				m.loading = true
				m.loadingMsg = "Rejecting..."
				m.joinRequestsMsg = ""
				return m, tea.Batch(resolveJoinRequestCmd(m.client, tid, req.ID, "rejected"), spinnerTickCmd())
			}
			return m, nil
		}
	}
	return m, nil
}

func (m Model) inviteCodesUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateProfile
			m.inviteCodes = nil
			m.inviteCodeCreated = ""
			m.inviteCodesMsg = ""
			return m, nil
		case "up", "k":
			if m.inviteCodesCursor > 0 {
				m.inviteCodesCursor--
			}
			return m, nil
		case "down":
			if m.inviteCodes != nil && m.inviteCodesCursor < len(m.inviteCodes.Data)-1 {
				m.inviteCodesCursor++
			}
			return m, nil
		case "c":
			m.loading = true
			m.loadingMsg = "Creating invite code..."
			m.inviteCodesMsg = ""
			return m, tea.Batch(createInviteCodeCmd(m.client), spinnerTickCmd())
		case "x":
			if m.inviteCodes != nil && len(m.inviteCodes.Data) > 0 {
				code := m.inviteCodes.Data[m.inviteCodesCursor]
				if code.Status == "active" {
					m.loading = true
					m.loadingMsg = "Invalidating..."
					m.inviteCodesMsg = ""
					return m, tea.Batch(invalidateInviteCodeCmd(m.client, code.ID), spinnerTickCmd())
				}
			}
			return m, nil
		}
	}
	return m, nil
}
