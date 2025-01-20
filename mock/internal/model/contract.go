package model

import (
	"errors"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var (
	ErrNotFound = errors.New("No such record found")
	ErrConflict = errors.New("Record already exists")
)

type DataAccess struct {
	Tokens TokensAccess
	Groups GroupsAccess
	Users  UsersAccess
}

type TokensAccess interface {
	Create(token string, uid int) error
	FindUserIdentifier(token string) (int, error)
}

type UsersAccess interface {
	Create(*gitlab.User) error
	FindByIdentifier(uid int) (*gitlab.User, error)
}

type GroupsAccess interface {
	Create(*gitlab.Group) error
	CreateUserAssociation(gid, uid int) error
	FindByUserIdentifier(uid, offset, size int) ([]*gitlab.Group, error)
	CountByUserIdentifier(uid int) (int, error)
}
