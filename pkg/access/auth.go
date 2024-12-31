package access

import (
	userauthz "github.com/UiP9AV6Y/go-k8s-user-authz"
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

// NewRejectDormantAuthorizer returns an [userauthz.Authorizer]
// which rejects users whose extra values contain [AttributeDormant]
func NewRejectDormantAuthorizer() userauthz.Authorizer {
	return userinfo.RejectExtra(GitlabAttributesKey, AttributeDormant)
}

// NewRequireUsersAuthorizer returns an [userauthz.Authorizer] instance
// which requires a user to be named in the given list.
func NewRequireUsersAuthorizer(users []string) userauthz.Authorizer {
	return userinfo.RequireAnyNames(users)
}

// NewRequireGroupsAuthorizer returns an [userauthz.Authorizer] instance
// which requires a user to be a member of ALL given groups.
func NewRequireGroupsAuthorizer(groups []string) userauthz.Authorizer {
	return userinfo.RequireAllGroups(groups)
}

// NewRejectUsersAuthorizer returns an [userauthz.Authorizer] instance
// which rejects a user if named in the given list.
func NewRejectUsersAuthorizer(users []string) userauthz.Authorizer {
	return userinfo.RejectAnyNames(users)
}

// NewRejectGroupsAuthorizer returns an [userauthz.Authorizer] instance
// which rejects users with membership of at least on of the given groups.
func NewRejectGroupsAuthorizer(groups []string) userauthz.Authorizer {
	return userinfo.RejectAnyGroups(groups)
}
