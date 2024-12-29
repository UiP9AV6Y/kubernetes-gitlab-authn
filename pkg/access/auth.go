package access

import (
	"context"
	"slices"

	"k8s.io/apiserver/pkg/authentication/user"

	"github.com/UiP9AV6Y/go-k8s-user-authz"
	"github.com/UiP9AV6Y/go-k8s-user-authz/userinfo"
)

// NewRequire2FAAuthorizer returns an [userauthz.Authorizer]
// which rejects users whose extra values DO NOT contain [Attribute2fa]
func NewRequire2FAAuthorizer() userauthz.Authorizer {
	return userinfo.RequireExtra(GitlabAttributesKey, Attribute2fa)
}

// NewRejectBotAuthorizer returns an [userauthz.Authorizer]
// which rejects users whose extra values contain [AttributeBot]
func NewRejectBotAuthorizer() userauthz.Authorizer {
	return userinfo.RejectExtra(GitlabAttributesKey, AttributeBot)
}

// NewRejectLockedAuthorizer returns an [userauthz.Authorizer]
// which rejects users whose extra values contain [AttributeLocked]
func NewRejectLockedAuthorizer() userauthz.Authorizer {
	return userinfo.RejectExtra(GitlabAttributesKey, AttributeLocked)
}

// NewRejectPristineAuthorizer returns an [userauthz.Authorizer]
// which rejects users whose extra values contain [AttributePristine]
func NewRejectPristineAuthorizer() userauthz.Authorizer {
	return userinfo.RejectExtra(GitlabAttributesKey, AttributePristine)
}

// NewRequireUsersAuthorizer returns an [userauthz.Authorizer] instance
// which requires a user to be named in the given list.
func NewRequireUsersAuthorizer(users []string) userauthz.Authorizer {
	auth := make([]userauthz.Authorizer, len(users))
	for i, u := range users {
		auth[i] = userinfo.RequireName(u)
	}

	return userauthz.RequireAny(auth)
}

// NewRequireGroupsAuthorizer returns an [userauthz.Authorizer] instance
// which requires a user to be a member of ALL given groups.
func NewRequireGroupsAuthorizer(groups []string) userauthz.Authorizer {
	auth := make([]userauthz.Authorizer, len(groups))
	for i, g := range groups {
		auth[i] = userinfo.RequireGroup(g)
	}

	return userauthz.RequireAll(auth)
}

// NewRejectUsersAuthorizer returns an [userauthz.Authorizer] instance
// which rejects a user if named in the given list.
func NewRejectUsersAuthorizer(users []string) userauthz.Authorizer {
	auth := func(_ context.Context, u user.Info) userauthz.Decision {
		name := u.GetName()
		if slices.Contains(users, name) {
			return userauthz.Decision("User " + name + " is explicitly prohibited")
		}

		return userauthz.DecisionAllow
	}

	return userauthz.AuthorizerFunc(auth)
}

// NewRejectGroupsAuthorizer returns an [userauthz.Authorizer] instance
// which rejects users with membership of at least on of the given groups.
func NewRejectGroupsAuthorizer(groups []string) userauthz.Authorizer {
	auth := func(_ context.Context, u user.Info) userauthz.Decision {
		haystack := u.GetGroups()
		if haystack == nil {
			return userauthz.DecisionAllow
		}

		for _, g := range haystack {
			if slices.Contains(haystack, g) {
				return userauthz.Decision("Members of " + g + " are not allowed")
			}
		}

		return userauthz.DecisionAllow
	}

	return userauthz.AuthorizerFunc(auth)
}
