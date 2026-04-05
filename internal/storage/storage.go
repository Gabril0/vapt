package storage

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
)

type BestResult struct {
	WPM      float64 `json:"wpm"`
	Accuracy float64 `json:"accuracy"`
	Score    float64 `json:"score"`
	Duration int     `json:"duration"`
}

const obfuscationKey = "instead-of-just-doomscrolling-try-to-improve-your-typing"

func bestFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "best.dat"
	}
	dir := filepath.Join(home, ".vapt")
	os.MkdirAll(dir, 0700)
	return filepath.Join(dir, "best.dat")
}

func obfuscate(data []byte) []byte {
	key := []byte(obfuscationKey)
	out := make([]byte, len(data))
	for i, b := range data {
		out[i] = b ^ key[i%len(key)]
	}
	return out
}

func LoadBest() BestResult {
	raw, err := os.ReadFile(bestFilePath())
	if err != nil {
		return BestResult{}
	}
	encrypted, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		return BestResult{}
	}
	data := obfuscate(encrypted)
	var b BestResult
	if err := json.Unmarshal(data, &b); err != nil {
		return BestResult{}
	}
	return b
}

func SaveBest(b BestResult) {
	data, err := json.Marshal(b)
	if err != nil {
		return
	}
	encrypted := obfuscate(data)
	encoded := base64.StdEncoding.EncodeToString(encrypted)
	os.WriteFile(bestFilePath(), []byte(encoded), 0600)
}
