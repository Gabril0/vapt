package api

type UserInfo struct {
	ID        string       `json:"id"`
	Email     string       `json:"email"`
	Username  string       `json:"username"`
	Role      string       `json:"role"`
	CreatedAt string       `json:"createdAt"`
	Tenants   []UserTenant `json:"tenants,omitempty"`
}

type UserTenant struct {
	TenantID string `json:"tenantId"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	JoinedAt string `json:"joinedAt"`
}

type LeaderboardData struct {
	Metric   string             `json:"metric"`
	Duration string             `json:"duration"`
	Period   string             `json:"period"`
	Entries  []LeaderboardEntry `json:"entries"`
}

type LeaderboardEntry struct {
	Rank         int             `json:"rank"`
	User         LeaderboardUser `json:"user"`
	BestValue    float64         `json:"bestValue"`
	WPM          float64         `json:"wpm"`
	Accuracy     float64         `json:"accuracy"`
	Consistency  float64         `json:"consistency"`
	TestDuration int             `json:"testDuration"`
	PlayedAt     string          `json:"playedAt"`
}

type LeaderboardUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type TenantData struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	MemberCount int            `json:"memberCount"`
	CreatedAt   string         `json:"createdAt"`
	Members     []TenantMember `json:"-"`
}

type TenantMember struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	JoinedAt string `json:"joinedAt"`
}

type SubmitResultReq struct {
	WPM          float64 `json:"wpm"`
	Accuracy     float64 `json:"accuracy"`
	Consistency  float64 `json:"consistency"`
	TestDuration int     `json:"testDuration"`
}

type SubmitResultResp struct {
	ID           string  `json:"id"`
	UserID       string  `json:"userId"`
	TenantID     string  `json:"tenantId"`
	WPM          float64 `json:"wpm"`
	Accuracy     float64 `json:"accuracy"`
	Consistency  float64 `json:"consistency"`
	TestDuration int     `json:"testDuration"`
	PlayedAt     string  `json:"playedAt"`
}

type PaginatedMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type TenantListEntry struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"memberCount"`
	CreatedAt   string `json:"createdAt"`
}

type PaginatedTenants struct {
	Data []TenantListEntry `json:"data"`
	Meta PaginatedMeta     `json:"meta"`
}

type JoinRequest struct {
	ID   string `json:"id"`
	User struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type PaginatedJoinRequests struct {
	Data []JoinRequest `json:"data"`
	Meta PaginatedMeta `json:"meta"`
}

type InviteCode struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	CreatedBy struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"createdBy"`
	UsedBy *struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"usedBy"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type PaginatedInviteCodes struct {
	Data []InviteCode  `json:"data"`
	Meta PaginatedMeta `json:"meta"`
}

type GameResult struct {
	ID           string  `json:"id"`
	WPM          float64 `json:"wpm"`
	Accuracy     float64 `json:"accuracy"`
	Consistency  float64 `json:"consistency"`
	TestDuration int     `json:"testDuration"`
	PlayedAt     string  `json:"playedAt"`
}

type PaginatedResults struct {
	Data []GameResult  `json:"data"`
	Meta PaginatedMeta `json:"meta"`
}

type UpdateProfileReq struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}
