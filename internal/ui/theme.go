package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Theme struct {
	Accent  string
	Text    string
	TextDim string
	Correct string
	Error   string
	ErrorBg string
	Cursor  string
	Warning string
	Success string
	Info    string
	Star    string
	Border  string
}

var CurrentTheme = LoadTheme()

func DefaultTheme() Theme {
	return Theme{
		Accent:  FgCyan,
		Text:    FgWhite,
		TextDim: FgGray,
		Correct: FgGreen,
		Error:   FgRed,
		ErrorBg: BgRed,
		Cursor:  FgWhite,
		Warning: FgYellow,
		Success: FgGreen,
		Info:    FgCyan,
		Star:    FgYellow,
		Border:  FgGray,
	}
}

func LoadTheme() Theme {
	t := DefaultTheme()

	data, err := os.ReadFile("theme.json")
	if err != nil {
		if exe, exeErr := os.Executable(); exeErr == nil {
			data, err = os.ReadFile(filepath.Join(filepath.Dir(exe), "theme.json"))
		}
		if err != nil {
			return t
		}
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return t
	}

	if v, ok := raw["accent"]; ok {
		t.Accent = parseColor(v, false)
	}
	if v, ok := raw["text"]; ok {
		t.Text = parseColor(v, false)
	}
	if v, ok := raw["text_dim"]; ok {
		t.TextDim = parseColor(v, false)
	}
	if v, ok := raw["correct"]; ok {
		t.Correct = parseColor(v, false)
	}
	if v, ok := raw["error"]; ok {
		t.Error = parseColor(v, false)
	}
	if v, ok := raw["error_bg"]; ok {
		t.ErrorBg = parseColor(v, true)
	}
	if v, ok := raw["cursor"]; ok {
		t.Cursor = parseColor(v, false)
	}
	if v, ok := raw["warning"]; ok {
		t.Warning = parseColor(v, false)
	}
	if v, ok := raw["success"]; ok {
		t.Success = parseColor(v, false)
	}
	if v, ok := raw["info"]; ok {
		t.Info = parseColor(v, false)
	}
	if v, ok := raw["star"]; ok {
		t.Star = parseColor(v, false)
	}
	if v, ok := raw["border"]; ok {
		t.Border = parseColor(v, false)
	}

	return t
}

func parseColor(val string, bg bool) string {
	val = strings.TrimSpace(val)

	named := map[string][2]string{
		"black":   {FgBlack, "\033[40m"},
		"red":     {FgRed, BgRed},
		"green":   {FgGreen, BgGreen},
		"yellow":  {FgYellow, "\033[43m"},
		"blue":    {FgBlue, "\033[44m"},
		"magenta": {FgMagenta, "\033[45m"},
		"cyan":    {FgCyan, "\033[46m"},
		"white":   {FgWhite, "\033[47m"},
		"gray":    {FgGray, "\033[100m"},
	}
	if pair, ok := named[strings.ToLower(val)]; ok {
		if bg {
			return pair[1]
		}
		return pair[0]
	}

	if strings.HasPrefix(val, "#") && len(val) == 7 {
		r, err1 := strconv.ParseUint(val[1:3], 16, 8)
		g, err2 := strconv.ParseUint(val[3:5], 16, 8)
		b, err3 := strconv.ParseUint(val[5:7], 16, 8)
		if err1 == nil && err2 == nil && err3 == nil {
			if bg {
				return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
			}
			return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
		}
	}

	if strings.HasPrefix(strings.ToLower(val), "rgb(") && strings.HasSuffix(val, ")") {
		inner := val[4 : len(val)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 3 {
			r, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			g, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			b, err3 := strconv.Atoi(strings.TrimSpace(parts[2]))
			if err1 == nil && err2 == nil && err3 == nil &&
				r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
				if bg {
					return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
				}
				return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
			}
		}
	}

	if strings.Contains(val, ";") {
		code := val
		if !strings.HasPrefix(code, "\033[") {
			code = "\033[" + code + "m"
		}
		return code
	}

	if bg {
		return BgRed
	}
	return FgWhite
}
