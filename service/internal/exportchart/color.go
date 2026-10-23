package exportchart

import (
	"image/color"
	"regexp"
	"strconv"
	"strings"
)

var hexColor = regexp.MustCompile(`#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})\b`)

var named = map[string]color.RGBA{
	"purple": {128, 0, 128, 255},
	"teal":   {0, 128, 128, 255},
	"red":    {180, 40, 40, 255},
	"navy":   {0, 0, 128, 255},
	"black":  {20, 20, 20, 255},
}

func accentColor(css string) color.RGBA {
	if match := hexColor.FindStringSubmatch(css); len(match) == 2 {
		return parseHex(match[1])
	}

	lower := strings.ToLower(css)
	for name, value := range named {
		if strings.Contains(lower, name) {
			return value
		}
	}

	return named["purple"]
}

func parseHex(hex string) color.RGBA {
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}

	n, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return named["purple"]
	}

	return color.RGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 255}
}
