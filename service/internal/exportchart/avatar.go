package exportchart

import (
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesread/orgviz/internal/orgs"

	_ "image/jpeg"
	_ "image/png"
)

func loadAvatar(dir string, person *orgs.Person) image.Image {
	path := avatarPath(dir, person)
	if path == "" {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil
	}

	return img
}

func avatarPath(dir string, person *orgs.Person) string {
	if dir == "" {
		return ""
	}

	if person.Avatar != "" {
		return filepath.Join(dir, filepath.Base(person.Avatar))
	}

	return guessedAvatar(dir, person.Name)
}

func guessedAvatar(dir, name string) string {
	fields := strings.Fields(name)
	if len(fields) == 0 {
		return filepath.Join(dir, "default.png")
	}

	guess := filepath.Join(dir, strings.ToLower(fields[0])+".jpg")
	if _, err := os.Stat(guess); err == nil {
		return guess
	}

	return filepath.Join(dir, "default.png")
}
