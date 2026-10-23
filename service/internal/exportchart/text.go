package exportchart

import "strings"

func wrap(text string, maxChars int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	lines := []string{words[0]}
	for _, word := range words[1:] {
		lines = appendWord(lines, word, maxChars)
	}

	return lines
}

func appendWord(lines []string, word string, maxChars int) []string {
	last := lines[len(lines)-1]
	if len(last)+1+len(word) <= maxChars {
		lines[len(lines)-1] = last + " " + word
		return lines
	}

	return append(lines, word)
}
