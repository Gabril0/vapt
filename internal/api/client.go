package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

func Is401(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return false
}

func (c *Client) doJSON(method, path string, body interface{}, out interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("server unavailable")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		msg := string(respBody)
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &errResp) == nil {
			if errResp.Error != "" {
				msg = errResp.Error
			} else if errResp.Message != "" {
				msg = errResp.Message
			}
		}
		return &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

type loginResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

func (c *Client) Login(username, password string) (*UserInfo, string, error) {
	var resp loginResponse
	err := c.doJSON("POST", "/auth/login", map[string]string{
		"email":    username,
		"password": password,
	}, &resp)
	if err != nil {
		return nil, "", err
	}
	return &resp.User, resp.Token, nil
}

func (c *Client) Register(email, username, password, inviteCode string) (*UserInfo, string, error) {
	body := map[string]string{
		"email":      email,
		"username":   username,
		"password":   password,
		"inviteCode": inviteCode,
	}
	var resp loginResponse
	err := c.doJSON("POST", "/auth/register", body, &resp)
	if err != nil {
		return nil, "", err
	}
	return &resp.User, resp.Token, nil
}

func (c *Client) GetTenant(tenantID string) (*TenantData, error) {
	var resp TenantData
	if err := c.doJSON("GET", "/tenants/"+tenantID, nil, &resp); err != nil {
		return nil, err
	}
	var membersResp struct {
		Data []TenantMember `json:"data"`
	}
	if err := c.doJSON("GET", "/tenants/"+tenantID+"/members", nil, &membersResp); err != nil {
		return nil, err
	}
	resp.Members = membersResp.Data
	return &resp, nil
}

func (c *Client) SubmitResult(tenantID string, req SubmitResultReq) (*SubmitResultResp, error) {
	var resp SubmitResultResp
	if err := c.doJSON("POST", "/tenants/"+tenantID+"/results", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetLeaderboard(tenantID, metric, duration, period string, limit int) (*LeaderboardData, error) {
	path := fmt.Sprintf("/tenants/%s/leaderboard?metric=%s&duration=%s&period=%s&limit=%d",
		tenantID, metric, duration, period, limit)
	var resp LeaderboardData
	if err := c.doJSON("GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ValidateToken() (*UserInfo, error) {
	var resp UserInfo
	err := c.doJSON("GET", "/users/me", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Logout() error {
	return c.doJSON("POST", "/auth/logout", nil, nil)
}

func (c *Client) UpdateProfile(req UpdateProfileReq) (*UserInfo, error) {
	var resp UserInfo
	if err := c.doJSON("PATCH", "/users/me", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ListTenants(search string, page, limit int) (*PaginatedTenants, error) {
	path := fmt.Sprintf("/tenants?search=%s&page=%d&limit=%d", search, page, limit)
	var resp PaginatedTenants
	if err := c.doJSON("GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CreateTenant(name string) (*TenantListEntry, error) {
	var resp TenantListEntry
	if err := c.doJSON("POST", "/tenants", map[string]string{"name": name}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) SendJoinRequest(tenantID string) error {
	return c.doJSON("POST", "/tenants/"+tenantID+"/join-requests", nil, nil)
}

func (c *Client) ListJoinRequests(tenantID, status string, page, limit int) (*PaginatedJoinRequests, error) {
	path := fmt.Sprintf("/tenants/%s/join-requests?status=%s&page=%d&limit=%d", tenantID, status, page, limit)
	var resp PaginatedJoinRequests
	if err := c.doJSON("GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ResolveJoinRequest(tenantID, requestID, status string) error {
	return c.doJSON("PATCH", "/tenants/"+tenantID+"/join-requests/"+requestID, map[string]string{"status": status}, nil)
}

func (c *Client) CreateInviteCode() (*InviteCode, error) {
	var resp InviteCode
	if err := c.doJSON("POST", "/invite-codes", map[string]string{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ListInviteCodes(page, limit int) (*PaginatedInviteCodes, error) {
	path := fmt.Sprintf("/invite-codes?page=%d&limit=%d", page, limit)
	var resp PaginatedInviteCodes
	if err := c.doJSON("GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) InvalidateInviteCode(id string) error {
	return c.doJSON("PATCH", "/invite-codes/"+id+"/invalidate", nil, nil)
}

func (c *Client) GetMyResults(tenantID string, page, limit int, duration, sort, order string) (*PaginatedResults, error) {
	path := fmt.Sprintf("/tenants/%s/results/me?page=%d&limit=%d&duration=%s&sort=%s&order=%s",
		tenantID, page, limit, duration, sort, order)
	var resp PaginatedResults
	if err := c.doJSON("GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
