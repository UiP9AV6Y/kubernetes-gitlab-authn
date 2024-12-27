package access

import (
	"slices"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const (
	ExplainUserNameRequire    = "Username is not explicitly allowed"
	ExplainUserNameReject     = "Username is explicitly rejected"
	ExplainUser2FARequire     = "Second factor authentication must be enabled"
	ExplainUserBotReject      = "Robot users are not allowed"
	ExplainUserLockedReject   = "User is locked"
	ExplainUserPristineReject = "User has never logged in before"
)

type UserRuler interface {
	Explain() string
	AuthorizeUser(*gitlab.User) bool
}

type UserRulerFunc func(u *gitlab.User) bool

var (
	User2FARequirement    UserRuler
	UserBotRejection      UserRuler
	UserLockedRejection   UserRuler
	UserPristineRejection UserRuler
)

type userRule struct {
	e string
	u UserRulerFunc
}

func (a *userRule) Explain() string {
	return a.e
}

func (a *userRule) AuthorizeUser(u *gitlab.User) bool {
	return a.u(u)
}

func newUserNameRule(n []string, allow bool, exp string) *userRule {
	logic := func(u *gitlab.User) bool {
		return allow == slices.Contains(n, u.Username)
	}
	result := &userRule{
		e: exp,
		u: logic,
	}

	return result
}

func UserNameRequirement(n []string) UserRuler {
	return newUserNameRule(n, true, ExplainUserNameRequire)
}

func UserNameRejection(n []string) UserRuler {
	return newUserNameRule(n, false, ExplainUserNameReject)
}

func init() {
	User2FARequirement = &userRule{
		e: ExplainUser2FARequire,
		u: func(u *gitlab.User) bool {
			return u.TwoFactorEnabled
		},
	}
	UserBotRejection = &userRule{
		e: ExplainUserBotReject,
		u: func(u *gitlab.User) bool {
			return u.Bot
		},
	}
	UserLockedRejection = &userRule{
		e: ExplainUserLockedReject,
		u: func(u *gitlab.User) bool {
			return !u.Locked
		},
	}
	UserPristineRejection = &userRule{
		e: ExplainUserPristineReject,
		u: func(u *gitlab.User) bool {
			return u.ConfirmedAt != nil
		},
	}
}
