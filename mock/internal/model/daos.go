package model

import (
	"strings"

	"gitlab.com/gitlab-org/api/client-go"
)

// primary keys for DAO usage
var (
	pkAdmin    = "ADMIN"
	pkMock     = "MOCK"
	pkSecure   = "SECURE"
	pkPrivate  = "PRIVATE"
	pkExternal = "EXTERNAl"
	pkBot      = "BOT"
	pkLocked   = "LOCKED"
	pkPristine = "PRISTINE"
	pkDormant  = "DORMANT"
)

// primitive data access objects
var (
	userDAO  map[string]gitlab.User
	groupDAO map[string][]gitlab.Group
)

type SelectUserQuery func(string) (gitlab.User, bool)

type SelectGroupsQuery func(string) ([]gitlab.Group, bool)

func SelectUserByTokenQuery(strict bool) SelectUserQuery {
	var q func(string) string
	if strict {
		q = findPKStrict
	} else {
		q = findPKLenient
	}

	return func(s string) (u gitlab.User, ok bool) {
		u, ok = userDAO[q(s)]

		return
	}
}

func SelectGroupsByTokenQuery(strict bool) SelectGroupsQuery {
	var q func(string) string
	if strict {
		q = findPKStrict
	} else {
		q = findPKLenient
	}

	return func(s string) (g []gitlab.Group, ok bool) {
		g, ok = groupDAO[q(s)]

		return
	}
}

func findPKLenient(s string) string {
	for pk, _ := range userDAO {
		if strings.Contains(s, pk) {
			return pk
		}
	}

	return ""
}

func findPKStrict(s string) string {
	if _, ok := userDAO[s]; ok {
		return s
	}

	return ""
}

func init() {
	userDAO = map[string]gitlab.User{
		pkAdmin:    adminUser,
		pkMock:     mockUser,
		pkSecure:   secureUser,
		pkPrivate:  privateUser,
		pkExternal: externalUser,
		pkBot:      botUser,
		pkLocked:   lockedUser,
		pkPristine: pristineUser,
		pkDormant:  dormantUser,
	}

	adminGroups := []gitlab.Group{
		coreGroup, adminGroup,
	}
	userGroups := []gitlab.Group{
		coreGroup, userGroup,
	}
	specialGroups := []gitlab.Group{
		coreGroup, userGroup, specialGroup,
	}
	groupDAO = map[string][]gitlab.Group{
		pkAdmin:    adminGroups,
		pkMock:     userGroups,
		pkSecure:   userGroups,
		pkPrivate:  userGroups,
		pkExternal: specialGroups,
		pkBot:      specialGroups,
		pkLocked:   specialGroups,
		pkPristine: specialGroups,
		pkDormant:  specialGroups,
	}
}
