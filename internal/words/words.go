package words

import (
	"bufio"
	"math/rand"
	"os"
	"regexp"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

var symbolRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]`)

func cleanText(s string) string {
	s = symbolRegex.ReplaceAllString(s, "")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}

var customWords = loadCustomWords()

func loadCustomWords() []string {
	file, err := os.Open("words.txt")
	if err != nil {
		return nil
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			words = append(words, line)
		}
	}
	if len(words) == 0 {
		return nil
	}
	return words
}

func GenerateText(minLen int) string {
	if customWords != nil {
		return generateFromCustom(minLen)
	}
	return generateFromFakeit(minLen)
}

func generateFromCustom(minLen int) string {
	var sb strings.Builder
	for sb.Len() < minLen {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(customWords[rand.Intn(len(customWords))])
	}
	return sb.String()
}

func generateFromFakeit(minLen int) string {
	var sb strings.Builder
	faker := gofakeit.New(uint64(rand.Int63()))
	for sb.Len() < minLen {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		var text string
		switch rand.Intn(3) {
		case 0:
			text = faker.Sentence(rand.Intn(8) + 3)
		case 1:
			text = faker.Phrase()
		case 2:
			text = faker.Question()
		}
		text = strings.TrimRight(strings.ToLower(text), ".?!")
		text = cleanText(text)
		if text != "" {
			sb.WriteString(text)
		}
	}
	return sb.String()
}
