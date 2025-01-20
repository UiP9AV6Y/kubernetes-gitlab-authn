package memory

import (
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock/internal/model"
)

type tokensAccess struct {
	s *storage
}

func (t *tokensAccess) Create(token string, uid int) error {
	if _, ok := t.s.tokenUsers[token]; ok {
		return model.ErrConflict
	}

	t.s.tokenUsers[token] = uid

	return nil
}

func (t *tokensAccess) FindUserIdentifier(token string) (int, error) {
	uid, ok := t.s.tokenUsers[token]
	if !ok {
		return -1, model.ErrNotFound
	}

	return uid, nil
}
