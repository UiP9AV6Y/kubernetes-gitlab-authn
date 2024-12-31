package model

import (
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
	privateUser = gitlab.User{
		ID:               4321,
		Username:         "private",
		Email:            "private@localhost.localdomain",
		Name:             "Private User",
		Provider:         userProvider,
		CustomAttributes: customUserAttributes,
		ConfirmedAt:      &confirmed,
		CreatedAt:        &created,
		PrivateProfile:   true,
	}
	externalUser = gitlab.User{
		ID:               5050,
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
		ID:               6443,
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
		ID:               7890,
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
		ID:               8080,
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
