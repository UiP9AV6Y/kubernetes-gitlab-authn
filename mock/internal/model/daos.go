package model

import (
	"strconv"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"
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

type SelectGroupsQuery func(string, int, int) ([]gitlab.Group, int, bool)

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

	return func(s string, offset, limit int) (groups []gitlab.Group, total int, ok bool) {
		high := offset + limit
		groups, ok = groupDAO[q(s)]
		if !ok || offset < 0 || limit <= 0 {
			return
		}

		total = len(groups)
		if offset >= total {
			groups = []gitlab.Group{}
		} else if high >= total {
			groups = groups[offset:]
		} else {
			groups = groups[offset:high]
		}

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

	featureSize := 50
	adminGroups := []gitlab.Group{
		coreGroup, adminGroup,
	}
	userGroups := make([]gitlab.Group, 0, featureSize+3)
	userGroups = append(userGroups, coreGroup, testGroup, userGroup)
	specialGroups := []gitlab.Group{
		coreGroup, userGroup, specialGroup,
	}
	for i := 1; i <= featureSize; i++ {
		feat := strconv.Itoa(i)
		path := "green-" + feat
		name := "Green " + feat
		if i%2 == 0 {
			path = "blue-" + feat
			name = "Blue " + feat
		}

		group := gitlab.Group{
			ID:          2000 + i,
			ParentID:    2,
			Description: "Feature selection pool",
			Name:        name,
			Path:        path,
			FullName:    "test / " + path,
			FullPath:    "test/" + path,
			CreatedAt:   &created,
			Visibility:  gitlab.PrivateVisibility,
		}
		userGroups = append(userGroups, group)
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
