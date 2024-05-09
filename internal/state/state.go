package state

import (
	"errors"
	"path"
)

const (
	_repoJSON    = "repo"
	_branchesDir = "branches"
)

type repoInfo struct {
	Trunk  string `json:"trunk"`
	Remote string `json:"remote"`
}

func (i *repoInfo) Validate() error {
	if i.Trunk == "" {
		return errors.New("trunk branch name is empty")
	}
	return nil
}

type branchStateBase struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

type branchState struct {
	Base branchStateBase `json:"base"`
	PR   int             `json:"pr,omitempty"`
}

// branchJSON returns the path to the JSON file for the given branch
// relative to the store's root.
func (s *Store) branchJSON(name string) string {
	return path.Join(_branchesDir, name)
}
