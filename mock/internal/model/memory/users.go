package memory

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock/internal/model"
)

type usersAccess struct {
	s *storage
}

func (u *usersAccess) Create(user *gitlab.User) error {
	if _, ok := u.s.users[user.ID]; ok {
		return model.ErrConflict
	}

	u.s.users[user.ID] = user

	return nil
}

func (u *usersAccess) FindByIdentifier(uid int) (*gitlab.User, error) {
	user, ok := u.s.users[uid]
	if !ok {
		return nil, model.ErrNotFound
	}

	return user, nil
}
