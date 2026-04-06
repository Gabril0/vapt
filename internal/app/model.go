package app

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"vapt/internal/api"
	"vapt/internal/auth"
	"vapt/internal/storage"
)

type state int

const (
	stateMenu state = iota
	stateTyping
	stateResults
	stateLogin
	stateRegister
	stateLeaderboard
	stateTenant
	stateProfile
	stateMyResults
	stateTenantList
	stateJoinRequests
	stateInviteCodes
)

var durations = []int{15, 30, 60, 120}

type Model struct {
	state        state
	target       string
	typed        string
	started      bool
	startTime    time.Time
	wpm          float64
	accuracy     float64
	consistency  float64
	score        float64
	duration     int
	durationIdx  int
	timeLeft     float64
	best         storage.BestResult
	wordTimes    []float64
	lastWordTime time.Time

	// Auth
	token   string
	user    *api.UserInfo
	offline bool
	client  *api.Client

	// Online features
	leaderboard *api.LeaderboardData
	tenant      *api.TenantData
	submitResp  *api.SubmitResultResp
	submitErr   string

	// Loading spinner
	spinnerIdx int

	// Leaderboard filters
	lbMetric   string
	lbDuration string
	lbPeriod   string

	// Login form
	loginField int
	loginEmail string
	loginPass  string
	loginErr   string

	// Register form
	regField      int
	regEmail      string
	regUsername   string
	regPass       string
	regInviteCode string
	regErr        string

	// Loading states
	loading    bool
	loadingMsg string

	// Active tenant
	activeTenantIdx int

	// Profile form
	profileField    int
	profileUsername string
	profilePassword string
	profileMsg      string
	profileErr      string

	// My Results
	myResults       *api.PaginatedResults
	myResultsFilter string
	myResultsSort   string
	myResultsOrder  string
	myResultsPage   int

	// Tenant list
	tenantListData     *api.PaginatedTenants
	tenantListPage     int
	tenantListCursor   int
	tenantListMsg      string
	tenantListCreating bool
	tenantListName     string

	// Join requests
	joinRequests       *api.PaginatedJoinRequests
	joinRequestsCursor int
	joinRequestsMsg    string

	// Invite codes
	inviteCodes       *api.PaginatedInviteCodes
	inviteCodesCursor int
	inviteCodeCreated string
	inviteCodesMsg    string

	// Used to restore active tenant across restarts
	savedTenantID string

	// Logo animation
	animFrame int
}

func (m Model) activeTenantID() string {
	if m.user == nil || len(m.user.Tenants) == 0 {
		return ""
	}
	idx := m.activeTenantIdx
	if idx < 0 || idx >= len(m.user.Tenants) {
		idx = 0
	}
	return m.user.Tenants[idx].TenantID
}

func (m Model) activeTenantName() string {
	if m.user == nil || len(m.user.Tenants) == 0 {
		return ""
	}
	idx := m.activeTenantIdx
	if idx < 0 || idx >= len(m.user.Tenants) {
		idx = 0
	}
	return m.user.Tenants[idx].Name
}

func (m Model) activeTenantRole() string {
	if m.user == nil || len(m.user.Tenants) == 0 {
		return ""
	}
	idx := m.activeTenantIdx
	if idx < 0 || idx >= len(m.user.Tenants) {
		idx = 0
	}
	return m.user.Tenants[idx].Role
}

func (m Model) menuActions() []string {
	var actions []string
	actions = append(actions, "start")
	if !m.offline && m.user != nil && len(m.user.Tenants) > 0 {
		actions = append(actions, "my_results", "leaderboard", "tenant")
	}
	actions = append(actions, "duration")
	if !m.offline {
		actions = append(actions, "tenants")
		if m.user != nil && len(m.user.Tenants) > 1 {
			actions = append(actions, "switch_tenant")
		}
		actions = append(actions, "profile", "logout")
	}
	actions = append(actions, "quit")
	return actions
}

func (m Model) menuKeyIndex(key string) int {
	if len(key) == 0 {
		return -1
	}
	n := 0
	for _, ch := range key {
		if ch < '0' || ch > '9' {
			return -1
		}
		n = n*10 + int(ch-'0')
	}
	actions := m.menuActions()
	idx := n
	if idx < 0 || idx >= len(actions) {
		return -1
	}
	return idx
}

func (m *Model) resolveActiveTenant() {
	if m.user == nil || len(m.user.Tenants) == 0 {
		m.activeTenantIdx = 0
		return
	}
	if m.savedTenantID != "" {
		for i, t := range m.user.Tenants {
			if t.TenantID == m.savedTenantID {
				m.activeTenantIdx = i
				return
			}
		}
	}
	m.activeTenantIdx = 0
}

func InitialModel() Model {
	token := auth.LoadToken()
	baseURL := os.Getenv("VUPT_URL")
	if baseURL == "" {
		baseURL = "https://vupt-api.onrender.com"
	}
	client := api.NewClient(baseURL, token)

	startState := stateLogin
	if token != "" {
		startState = stateMenu
	}

	savedTenantID := auth.LoadActiveTenant()

	return Model{
		state:           state(startState),
		durationIdx:     0,
		duration:        durations[0],
		best:            storage.LoadBest(),
		token:           token,
		client:          client,
		offline:         token == "",
		lbMetric:        "score",
		lbDuration:      "all",
		lbPeriod:        "daily",
		myResultsFilter: "all",
		myResultsSort:   "played_at",
		myResultsOrder:  "desc",
		myResultsPage:   1,
		tenantListPage:  1,
		savedTenantID:   savedTenantID,
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{animTickCmd()}
	if m.token != "" {
		cmds = append(cmds, validateTokenCmd(m.client))
	}
	return tea.Batch(cmds...)
}
