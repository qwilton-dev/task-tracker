package domain

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var projectKeyRegex = regexp.MustCompile(`^[A-Z]{2,5}$`)

type Project struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	CreatedAt   time.Time `json:"created_at"`
}

var (
	ErrProjectNameRequired = errors.New("project name is required")
	ErrProjectKeyRequired  = errors.New("project key is required")
	ErrProjectKeyInvalid   = errors.New("project key must be 2-5 uppercase letters")
)

func NewProject(workspaceID, name, key string) (*Project, error) {
	if name == "" {
		return nil, ErrProjectNameRequired
	}
	if key == "" {
		return nil, ErrProjectKeyRequired
	}
	if !projectKeyRegex.MatchString(key) {
		return nil, ErrProjectKeyInvalid
	}
	return &Project{
		WorkspaceID: workspaceID,
		Name:        name,
		Key:         key,
	}, nil
}

func GenerateKey(name string) string {
	words := strings.Fields(strings.ToUpper(name))
	if len(words) == 1 {
		w := words[0]
		if len(w) >= 5 {
			return w[:5]
		}
		if len(w) >= 2 {
			return w
		}
		return w + "X"
	}
	key := ""
	for _, w := range words {
		if len(key) >= 5 {
			break
		}
		key += string(w[0])
	}
	if len(key) < 2 {
		key += "X"
	}
	return key
}

func GenerateUniqueKey(name string, exists func(key string) bool) string {
	key := GenerateKey(name)
	if !exists(key) {
		return key
	}
	for i := 2; i <= 99; i++ {
		suffix := strconv.Itoa(i)
		candidate := key
		if len(candidate)+len(suffix) > 5 {
			candidate = candidate[:5-len(suffix)]
		}
		candidate += suffix
		if !exists(candidate) {
			return candidate
		}
	}
	return key
}
