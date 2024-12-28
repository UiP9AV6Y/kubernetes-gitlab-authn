package access

import (
	"slices"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var (
	UserDefaultRequirement AccessRuler
	UserDefaultRejection   AccessRuler
	User2FARequirement     AccessRuler
	UserBotRejection       AccessRuler
	UserLockedRejection    AccessRuler
	UserPristineRejection  AccessRuler
)

func newUserNameRule(n []string, allow bool) AccessRuler {
	result := func(u *gitlab.User, _ []*gitlab.Group) bool {
		return allow == slices.Contains(n, u.Username)
	}

	return AccessRulerFunc(result)
}

func UserNameRequirement(n []string) AccessRuler {
	return newUserNameRule(n, true)
}

func UserNameRejection(n []string) AccessRuler {
	return newUserNameRule(n, false)
}

func init() {
	UserDefaultRequirement = AccessRulerFunc(func(_ *gitlab.User, _ []*gitlab.Group) bool {
		return true
	})
	UserDefaultRejection = AccessRulerFunc(func(_ *gitlab.User, _ []*gitlab.Group) bool {
		return false
	})
	User2FARequirement = AccessRulerFunc(func(u *gitlab.User, _ []*gitlab.Group) bool {
		return u.TwoFactorEnabled
	})
	UserBotRejection = AccessRulerFunc(func(u *gitlab.User, _ []*gitlab.Group) bool {
		return !u.Bot
	})
	UserLockedRejection = AccessRulerFunc(func(u *gitlab.User, _ []*gitlab.Group) bool {
		return !u.Locked
	})
	UserPristineRejection = AccessRulerFunc(func(u *gitlab.User, _ []*gitlab.Group) bool {
		return u.ConfirmedAt != nil
	})
}
