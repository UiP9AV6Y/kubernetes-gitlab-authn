package access

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func newGroupNameRule(n []string, allow, report bool) AccessRuler {
	result := func(_ *gitlab.User, groups []*gitlab.Group) bool {
		for _, needle := range n {
			for _, group := range groups {
				found := group.Path == needle
				if found == allow {
					return report
				}
			}
		}

		return !report
	}

	return AccessRulerFunc(result)
}

func AnyGroupNameRequirement(n []string) AccessRuler {
	return newGroupNameRule(n, true, true)
}

func AllGroupNameRequirement(n []string) AccessRuler {
	return newGroupNameRule(n, false, false)
}

func AnyGroupNameRejection(n []string) AccessRuler {
	return newGroupNameRule(n, true, false)
}

func AllGroupNameRejection(n []string) AccessRuler {
	return newGroupNameRule(n, false, true)
}
