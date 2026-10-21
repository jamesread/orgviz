package orgs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

// File is one organisation chart loaded from YAML.
type File struct {
	Title  string
	People []*Person

	peopleIDs map[string]int32
}

// Person is one row in an organisation file.
type Person struct {
	Name       string
	Alias      string
	Title      string
	Team       string
	Reports    string
	Attributes map[string]string

	ID          int32
	ReportsToID int32
}

// Parse unmarshals a single org YAML document and links reporting lines.
func Parse(contents []byte) (*File, error) {
	parsed := &File{}
	if err := yaml.Unmarshal(contents, parsed); err != nil {
		return nil, fmt.Errorf("parsing org file: %w", err)
	}

	assignPeopleIDs(parsed)
	linkPeopleIDs(parsed)

	return parsed, nil
}

// LoadDir reads every *.yml file in dir, in lexical order.
func LoadDir(dir string) ([]*File, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, fmt.Errorf("listing org files: %w", err)
	}

	orgs := make([]*File, 0, len(matches))
	for _, filename := range matches {
		org, err := loadFile(filename)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

func loadFile(filename string) (*File, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading org file %s: %w", filename, err)
	}

	org, err := Parse(contents)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	log.WithFields(log.Fields{
		"file":   filename,
		"title":  org.Title,
		"people": len(org.People),
	}).Info("Loaded org file")

	return org, nil
}

func assignPeopleIDs(org *File) {
	org.peopleIDs = make(map[string]int32)
	nextID := int32(1)

	for _, person := range org.People {
		nextID = assignPersonID(org, person, nextID)
	}
}

func assignPersonID(org *File, person *Person, nextID int32) int32 {
	if id, found := org.peopleIDs[person.Name]; found {
		person.ID = id
		return nextID
	}

	person.ID = nextID
	org.peopleIDs[person.Name] = nextID
	registerAlias(org, person, nextID)

	return nextID + 1
}

func registerAlias(org *File, person *Person, id int32) {
	if person.Alias == "" {
		return
	}

	org.peopleIDs[person.Alias] = id
}

func linkPeopleIDs(org *File) {
	for _, person := range org.People {
		if person.Reports == "" {
			continue
		}

		person.ReportsToID = org.peopleIDs[person.Reports]
	}
}

// AvatarURL is /avatars/<first-name>.jpg, using the first token of the name.
func AvatarURL(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	firstName := strings.Split(name, " ")[0]
	return "/avatars/" + strings.ToLower(firstName) + ".jpg"
}
