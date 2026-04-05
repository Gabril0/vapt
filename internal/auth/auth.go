package auth

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func tokenDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".vapt")
}

func tokenPath() string {
	dir := tokenDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "token")
}

func LoadToken() string {
	path := tokenPath()
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	token := strings.TrimSpace(string(data))
	if !IsValidToken(token) {
		return ""
	}
	return token
}

func SaveToken(token string) error {
	dir := tokenDir()
	if dir == "" {
		return os.ErrNotExist
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(tokenPath(), []byte(token), 0600)
}

func ClearToken() error {
	path := tokenPath()
	if path == "" {
		return nil
	}
	return os.Remove(path)
}

func activeTenantPath() string {
	dir := tokenDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "tenant")
}

func LoadActiveTenant() string {
	path := activeTenantPath()
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func SaveActiveTenant(tenantID string) error {
	dir := tokenDir()
	if dir == "" {
		return os.ErrNotExist
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(activeTenantPath(), []byte(tenantID), 0600)
}

func IsValidToken(token string) bool {
	if len(token) == 0 || len(token) > 2048 {
		return false
	}
	for _, r := range token {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func SanitizeInput(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if sb.Len() >= 64 {
			break
		}
		if r <= unicode.MaxASCII && unicode.IsPrint(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
