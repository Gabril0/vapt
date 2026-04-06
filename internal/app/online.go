package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"vapt/internal/api"
)

type validateTokenMsg struct {
	user *api.UserInfo
	err  error
}

type loginResultMsg struct {
	user  *api.UserInfo
	token string
	err   error
}

type registerResultMsg struct {
	user  *api.UserInfo
	token string
	err   error
}

type leaderboardMsg struct {
	data *api.LeaderboardData
	err  error
}

type tenantMsg struct {
	data *api.TenantData
	err  error
}

type submitResultMsg struct {
	resp *api.SubmitResultResp
	err  error
}

type logoutMsg struct {
	err error
}

type updateProfileMsg struct {
	user *api.UserInfo
	err  error
}

type myResultsMsg struct {
	data *api.PaginatedResults
	err  error
}

type tenantListMsg struct {
	data *api.PaginatedTenants
	err  error
}

type createTenantMsg struct {
	err error
}

type joinRequestsMsg struct {
	data *api.PaginatedJoinRequests
	err  error
}

type resolveJoinRequestMsg struct {
	err error
}

type sendJoinRequestMsg struct {
	err error
}

type inviteCodesMsg struct {
	data *api.PaginatedInviteCodes
	err  error
}

type createInviteCodeMsg struct {
	code *api.InviteCode
	err  error
}

type invalidateInviteCodeMsg struct {
	err error
}

func validateTokenCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		user, err := client.ValidateToken()
		return validateTokenMsg{user: user, err: err}
	}
}

func loginCmd(client *api.Client, username, password string) tea.Cmd {
	return func() tea.Msg {
		user, token, err := client.Login(username, password)
		return loginResultMsg{user: user, token: token, err: err}
	}
}

func registerCmd(client *api.Client, email, username, password, inviteCode string) tea.Cmd {
	return func() tea.Msg {
		user, token, err := client.Register(email, username, password, inviteCode)
		return registerResultMsg{user: user, token: token, err: err}
	}
}

func fetchLeaderboardCmd(client *api.Client, tenantID, metric, duration, period string, limit int) tea.Cmd {
	return func() tea.Msg {
		data, err := client.GetLeaderboard(tenantID, metric, duration, period, limit)
		return leaderboardMsg{data: data, err: err}
	}
}

func fetchTenantCmd(client *api.Client, tenantID string) tea.Cmd {
	return func() tea.Msg {
		data, err := client.GetTenant(tenantID)
		return tenantMsg{data: data, err: err}
	}
}

func submitResultCmd(client *api.Client, tenantID string, req api.SubmitResultReq) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.SubmitResult(tenantID, req)
		return submitResultMsg{resp: resp, err: err}
	}
}

func logoutCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		err := client.Logout()
		return logoutMsg{err: err}
	}
}

func updateProfileCmd(client *api.Client, req api.UpdateProfileReq) tea.Cmd {
	return func() tea.Msg {
		user, err := client.UpdateProfile(req)
		return updateProfileMsg{user: user, err: err}
	}
}

func fetchMyResultsCmd(client *api.Client, tenantID string, page, limit int, duration, sort, order string) tea.Cmd {
	return func() tea.Msg {
		data, err := client.GetMyResults(tenantID, page, limit, duration, sort, order)
		return myResultsMsg{data: data, err: err}
	}
}

func fetchTenantListCmd(client *api.Client, search string, page, limit int) tea.Cmd {
	return func() tea.Msg {
		data, err := client.ListTenants(search, page, limit)
		return tenantListMsg{data: data, err: err}
	}
}

func createTenantCmd(client *api.Client, name string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.CreateTenant(name)
		return createTenantMsg{err: err}
	}
}

func fetchJoinRequestsCmd(client *api.Client, tenantID, status string, page, limit int) tea.Cmd {
	return func() tea.Msg {
		data, err := client.ListJoinRequests(tenantID, status, page, limit)
		return joinRequestsMsg{data: data, err: err}
	}
}

func resolveJoinRequestCmd(client *api.Client, tenantID, requestID, status string) tea.Cmd {
	return func() tea.Msg {
		err := client.ResolveJoinRequest(tenantID, requestID, status)
		return resolveJoinRequestMsg{err: err}
	}
}

func sendJoinRequestCmd(client *api.Client, tenantID string) tea.Cmd {
	return func() tea.Msg {
		err := client.SendJoinRequest(tenantID)
		return sendJoinRequestMsg{err: err}
	}
}

func fetchInviteCodesCmd(client *api.Client, page, limit int) tea.Cmd {
	return func() tea.Msg {
		data, err := client.ListInviteCodes(page, limit)
		return inviteCodesMsg{data: data, err: err}
	}
}

func createInviteCodeCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		code, err := client.CreateInviteCode()
		return createInviteCodeMsg{code: code, err: err}
	}
}

func invalidateInviteCodeCmd(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		err := client.InvalidateInviteCode(id)
		return invalidateInviteCodeMsg{err: err}
	}
}
