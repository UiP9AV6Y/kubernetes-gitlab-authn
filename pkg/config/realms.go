package config

import (
	"github.com/UiP9AV6Y/go-k8s-user-authz"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/access"
)

type RealmAccessRules struct {
	// Reject users without 2FA set up
	Require2FA bool `json:"require_2fa"`
	// Reject users marked as robots
	RejectBots bool `json:"reject_bots"`
	// Reject users in locked state
	RejectLocked bool `json:"reject_locked"`
	// Reject users which have not confirmed their account yet
	RejectPristine bool `json:"reject_pristine"`
	// Only allow users with the given usernames
	RequireUsers []string `json:"require_users"`
	// Reject users based on their username
	RejectUsers []string `json:"reject_users"`
	// Require membership of all of these groups
	RequireGroups []string `json:"require_groups"`
	// Reject members of any of the given groups
	RejectGroups []string `json:"reject_groups"`
}

func (r *RealmAccessRules) UserRules() userauthz.Authorizer {
	result := []userauthz.Authorizer{}

	if r.Require2FA {
		result = append(result, access.NewRequire2FAAuthorizer())
	}

	if r.RejectBots {
		result = append(result, access.NewRejectBotAuthorizer())
	}

	if r.RejectLocked {
		result = append(result, access.NewRejectLockedAuthorizer())
	}

	if r.RejectPristine {
		result = append(result, access.NewRejectPristineAuthorizer())
	}

	if len(r.RequireUsers) > 0 {
		result = append(result, access.NewRequireUsersAuthorizer(r.RequireUsers))
	}

	if len(r.RequireGroups) > 0 {
		result = append(result, access.NewRequireGroupsAuthorizer(r.RequireGroups))
	}

	if len(r.RejectUsers) > 0 {
		result = append(result, access.NewRejectUsersAuthorizer(r.RejectUsers))
	}

	if len(r.RejectGroups) > 0 {
		result = append(result, access.NewRejectGroupsAuthorizer(r.RejectGroups))
	}

	return userauthz.RequireAll(result)
}

type RealmAccessList []*RealmAccessRules

func (r RealmAccessList) UserRules() userauthz.Authorizer {
	result := make([]userauthz.Authorizer, len(r))
	for i, u := range r {
		result[i] = u.UserRules()
	}

	return userauthz.RejectNoOpinion(
		userauthz.RequireAny(result),
		userauthz.Decision("No explicit permission"),
	)
}

type Realms map[string]RealmAccessList

func NewRealms() map[string]RealmAccessList {
	return map[string]RealmAccessList{}
}

func (r Realms) UserAccessControlList() map[string]userauthz.Authorizer {
	if len(r) == 0 {
		// allow anyone into the default realm
		// if nothing has been configured
		return map[string]userauthz.Authorizer{
			"": userauthz.AlwaysAllowAuthorizer,
		}
	}

	result := make(map[string]userauthz.Authorizer, len(r))
	for realm, acls := range r {
		result[realm] = acls.UserRules()
	}

	return result
}
