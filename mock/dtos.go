package main

import (
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var (
	// simulated time of confirmation
	confirmed = time.Now()
	// simulated time of creation
	created = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
)

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
		ConfirmedAt:      &confirmed,
		CreatedAt:        &created,
		IsAdmin:          true,
	}
	mockUser = gitlab.User{
		ID:               2468,
		Username:         "mock",
		Email:            "mock@localhost.localdomain",
		Name:             "Mock User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &confirmed,
		CreatedAt:        &created,
	}
	secureUser = gitlab.User{
		ID:               3456,
		Username:         "secure",
		Email:            "secure@localhost.localdomain",
		Name:             "Secure User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &confirmed,
		CreatedAt:        &created,
		TwoFactorEnabled: true,
	}
	externalUser = gitlab.User{
		ID:               4321,
		Username:         "external",
		Email:            "external@localhost.localdomain",
		Name:             "External User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &confirmed,
		CreatedAt:        &created,
		External:         true,
	}
	botUser = gitlab.User{
		ID:               5050,
		Username:         "bot",
		Email:            "bot@localhost.localdomain",
		Name:             "Bot User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &confirmed,
		CreatedAt:        &created,
		Bot:              true,
	}
	lockedUser = gitlab.User{
		ID:               6443,
		Username:         "locked",
		Email:            "locked@localhost.localdomain",
		Name:             "Locked User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &confirmed,
		CreatedAt:        &created,
		Locked:           true,
	}
	pristineUser = gitlab.User{
		ID:               7890,
		Username:         "pristine",
		Email:            "pristine@localhost.localdomain",
		Name:             "Pristine User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		CreatedAt:        &created,
	}
)

// group models
var (
	coreGroup = gitlab.Group{
		ID:         1,
		Name:       "Core",
		Path:       "core",
		FullName:   "core",
		FullPath:   "core",
		CreatedAt:  &created,
		Visibility: gitlab.PublicVisibility,
	}
	adminGroup = gitlab.Group{
		ID:          11,
		ParentID:    1,
		Description: "Operators, SREs, and maintenance staff",
		Name:        "Administrators",
		Path:        "admins",
		FullName:    "core / admins",
		FullPath:    "core/admins",
		CreatedAt:   &created,
		Visibility:  gitlab.PrivateVisibility,
	}
	userGroup = gitlab.Group{
		ID:          12,
		ParentID:    1,
		Description: "Regular users",
		Name:        "Users",
		Path:        "core-users",
		FullName:    "core / users",
		FullPath:    "core/users",
		CreatedAt:   &created,
		Visibility:  gitlab.InternalVisibility,
	}
	specialGroup = gitlab.Group{
		ID:          13,
		ParentID:    1,
		Description: "Users with one or more special attributes set to TRUE",
		Name:        "Non-conformant attributes",
		Path:        "special",
		FullName:    "core / special",
		FullPath:    "core/special",
		CreatedAt:   &created,
		Visibility:  gitlab.InternalVisibility,
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
