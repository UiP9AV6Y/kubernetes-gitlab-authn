package memory

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock/internal/model"
)

type storage struct {
	users      map[int]*gitlab.User
	groups     map[int]*gitlab.Group
	userGroups map[int][]int
	tokenUsers map[string]int
}

// NewDataAccess returns a cheap [model.DataAccess] implementation
// using various collections as storage backend. access is not synchronized,
// which makes this only suitable for read-only access.
func NewDataAccess() (*model.DataAccess, error) {
	s := &storage{
		users:      map[int]*gitlab.User{},
		groups:     map[int]*gitlab.Group{},
		userGroups: map[int][]int{},
		tokenUsers: map[string]int{},
	}
	tokens := &tokensAccess{
		s: s,
	}
	groups := &groupsAccess{
		s: s,
	}
	users := &usersAccess{
		s: s,
	}
	result := &model.DataAccess{
		Tokens: tokens,
		Groups: groups,
		Users:  users,
	}

	return result, nil
}
