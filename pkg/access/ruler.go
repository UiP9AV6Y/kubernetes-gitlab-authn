package access

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type AccessRuler interface {
	Authorize(*gitlab.User, []*gitlab.Group) bool
}

type AccessRulerFunc func(*gitlab.User, []*gitlab.Group) bool

func (f AccessRulerFunc) Authorize(u *gitlab.User, g []*gitlab.Group) bool {
	return f(u, g)
}

// AnyAccessRulers is a multi-rule evaluator which requires just
// a single requirement to succeed. It returns on the first match.
// It returns false if no rules apply.
type AnyAccessRulers []AccessRuler

func (a AnyAccessRulers) Authorize(user *gitlab.User, groups []*gitlab.Group) bool {
	for _, r := range a {
		if r.Authorize(user, groups) {
			return true
		}
	}

	return false
}

// AllAccessRulers is a multi-rule evaluator which requires for all
// requirements to be satisfied. It bails on the first violation.
// It returns true if no rules exist.
type AllAccessRulers []AccessRuler

func (a AllAccessRulers) Authorize(user *gitlab.User, groups []*gitlab.Group) bool {
	for _, r := range a {
		if !r.Authorize(user, groups) {
			return false
		}
	}

	return true
}
