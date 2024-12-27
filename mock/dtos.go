package main

import (
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// simulated time of confirmation
var now = time.Now()

// User provider property
var userProvider = "gitlab-mock"

// shared user attributes
var customUserAttributes = []*gitlab.CustomAttribute{
	&gitlab.CustomAttribute{
		Key:   "scope",
		Value: "testing-only",
	},
}

// user models
var (
	adminUser = gitlab.User{
		ID:               1234,
		Username:         "admin",
		Email:            "root@localhost.localdomain",
		Name:             "Admin User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &now,
		IsAdmin:          true,
	}
	mockUser = gitlab.User{
		ID:               2468,
		Username:         "mock",
		Email:            "mock@localhost.localdomain",
		Name:             "Mock User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &now,
	}
	secureUser = gitlab.User{
		ID:               3456,
		Username:         "secure",
		Email:            "secure@localhost.localdomain",
		Name:             "Secure User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &now,
		TwoFactorEnabled: true,
	}
	externalUser = gitlab.User{
		ID:               4321,
		Username:         "external",
		Email:            "external@localhost.localdomain",
		Name:             "External User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &now,
		External:         true,
	}
	botUser = gitlab.User{
		ID:               5050,
		Username:         "bot",
		Email:            "bot@localhost.localdomain",
		Name:             "Bot User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &now,
		Bot:              true,
	}
	lockedUser = gitlab.User{
		ID:               6443,
		Username:         "locked",
		Email:            "locked@localhost.localdomain",
		Name:             "Locked User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &now,
		Locked:           true,
	}
	pristineUser = gitlab.User{
		ID:               7890,
		Username:         "pristine",
		Email:            "pristine@localhost.localdomain",
		Name:             "Pristine User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
	}
)

// group models
var (
	coreGroup = gitlab.Group{
		ID:   1,
		Name: "Core",
		Path: "core",
	}
	adminGroup = gitlab.Group{
		ID:   		11,
		Name: 		"Administrators",
		Path: 		"core-admins",
		ParentID: 1,
	}
	userGroup = gitlab.Group{
		ID:   		12,
		Name: 		"Users",
		Path: 		"core-users",
		ParentID: 1,
	}
	specialGroup = gitlab.Group{
		ID:   		13,
		Name: 		"Non-conformant attributes",
		Path: 		"core-special",
		ParentID: 1,
	}
)

// primary keys for DAO usage
var (
	pkAdmin    = "ADMIN"
	pkMock     = "MOCK"
	pkSecure   = "SECURE"
	pkExternal = "EXTERNAl"
	pkBot      = "BOT"
	pkLocked   = "LOCKED"
	pkPristine = "PRISTINE"
	pks        = []string{
		pkAdmin,
		pkMock,
		pkSecure,
		pkExternal,
		pkBot,
		pkLocked,
		pkPristine,
	}
)

// primitive data access objects
var (
	userDAO  map[string]gitlab.User
	groupDAO map[string][]gitlab.Group
)

func init() {
	userDAO = make(map[string]gitlab.User, 6)
	userDAO[pkAdmin] = adminUser
	userDAO[pkMock] = mockUser
	userDAO[pkSecure] = secureUser
	userDAO[pkExternal] = externalUser
	userDAO[pkBot] = botUser
	userDAO[pkLocked] = lockedUser
	userDAO[pkPristine] = pristineUser

	adminGroups := []gitlab.Group{
		coreGroup, adminGroup,
	}
	userGroups := []gitlab.Group{
		coreGroup, userGroup,
	}
	specialGroups := []gitlab.Group{
		coreGroup, userGroup, specialGroup,
	}
	groupDAO = make(map[string][]gitlab.Group, 6)
	groupDAO[pkAdmin] = adminGroups
	groupDAO[pkMock] = userGroups
	groupDAO[pkSecure] = userGroups
	groupDAO[pkExternal] = specialGroups
	groupDAO[pkBot] = specialGroups
	groupDAO[pkLocked] = specialGroups
	groupDAO[pkPristine] = specialGroups
}

// findPK performs a substring search with the primary keys
// against the provided haystack.
func findPK(s string) string {
	for _, pk := range pks {
		if strings.Contains(s, pk) {
			return pk
		}
	}

	return ""
}
