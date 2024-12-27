package access

import (
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type UserRulers []UserRuler

func (rs UserRulers) AuthorizeUser(user *gitlab.User) bool {
	for _, r := range rs {
		if !r.AuthorizeUser(user) {
			return false
		}
	}

	return true
}

type UserRealmRuler map[string][]UserRuler

func NewDefaultUserRealmRuler(rules ...UserRuler) UserRealmRuler {
	result := map[string][]UserRuler{
		"": rules,
	}

	return result
}

func (a UserRealmRuler) AuthorizeUser(realm string, user *gitlab.User) error {
	err := a.authorize(realm, user)
	if err != nil {
		return err
	}

	_, hasDefaultRealm := a[""]
	if realm != "" && hasDefaultRealm {
		return a.authorize("", user)
	}

	return nil
}

func (a UserRealmRuler) authorize(realm string, user *gitlab.User) error {
	rules, ok := a[realm]
	if !ok {
		return fmt.Errorf("No such authentication realm %q", realm)
	}

	for _, r := range rules {
		if !r.AuthorizeUser(user) {
			return fmt.Errorf("user %q is not authorized to access realm %q: %s", user.Username, realm, r.Explain())
		}
	}

	return nil
}
