package access

import (
	"strconv"
	"strings"

	"gitlab.com/gitlab-org/api/client-go"

	authentication "k8s.io/api/authentication/v1"
)

func UserInfo(user *gitlab.User, groups []*gitlab.Group, attrGroups bool) authentication.UserInfo {
	var gids []string
	if attrGroups {
		agids := UserAttributeGroups(user)
		gids = make([]string, len(groups), len(groups)+len(agids))
		gids = append(gids, agids...)
	} else {
		gids = make([]string, len(groups))
	}

	for i, g := range groups {
		gids[i] = strings.ReplaceAll(g.FullPath, "/", ":")
	}

	extra := UserAttributeExtra(user)
	info := authentication.UserInfo{
		Username: user.Username,
		UID:      strconv.FormatInt(int64(user.ID), 10),
		Groups:   gids,
		Extra:    extra,
	}

	return info
}

func UserAttributeGroups(user *gitlab.User) []string {
	groups := make([]string, 0, 5)

	if user.TwoFactorEnabled {
		groups = append(groups, Group2fa)
	}
	if user.Bot {
		groups = append(groups, GroupBot)
	}
	if user.IsAdmin {
		groups = append(groups, GroupAdmin)
	}
	if user.IsAuditor {
		groups = append(groups, GroupAuditor)
	}
	if user.External {
		groups = append(groups, GroupExternal)
	}
	if user.PrivateProfile {
		groups = append(groups, GroupPrivate)
	}
	if user.Locked {
		groups = append(groups, GroupLocked)
	}
	if user.ConfirmedAt == nil {
		groups = append(groups, GroupPristine)
	}

	return groups
}

func UserAttributeExtra(user *gitlab.User) map[string]authentication.ExtraValue {
	attrs := make([]string, 0, 5)
	if user.TwoFactorEnabled {
		attrs = append(attrs, Attribute2fa)
	}
	if user.Bot {
		attrs = append(attrs, AttributeBot)
	}
	if user.IsAdmin {
		attrs = append(attrs, AttributeAdmin)
	}
	if user.IsAuditor {
		attrs = append(attrs, AttributeAuditor)
	}
	if user.External {
		attrs = append(attrs, AttributeExternal)
	}
	if user.PrivateProfile {
		attrs = append(attrs, AttributePrivate)
	}
	if user.Locked {
		attrs = append(attrs, AttributeLocked)
	}
	if user.ConfirmedAt == nil {
		attrs = append(attrs, AttributePristine)
	}

	extra := map[string]authentication.ExtraValue{
		GitlabAttributesKey: attrs,
	}

	for _, attr := range user.CustomAttributes {
		extra[GitlabKeyNamespace+attr.Key] = []string{attr.Value}
	}

	return extra
}
