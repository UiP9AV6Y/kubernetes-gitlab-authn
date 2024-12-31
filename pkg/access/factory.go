package access

import (
	"strconv"
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	authentication "k8s.io/api/authentication/v1"
)

type UserInfoOptions struct {
	AttributesAsGroups bool
	DormantTimeout     time.Duration
	Now                func() time.Time
}

func UserInfo(user *gitlab.User, groups []*gitlab.Group, opts UserInfoOptions) authentication.UserInfo {
	var gids []string
	var dormant bool
	var now time.Time

	if opts.Now != nil {
		now = opts.Now()
	} else {
		now = time.Now()
	}

	if user.LastActivityOn != nil && opts.DormantTimeout > 0 {
		dormant = now.Add(-opts.DormantTimeout).After(time.Time(*user.LastActivityOn))
	}

	if opts.AttributesAsGroups {
		agids := userAttributeGroups(user, dormant)
		gids = make([]string, len(groups), len(groups)+len(agids))
		gids = append(gids, agids...)
	} else {
		gids = make([]string, len(groups))
	}

	for i, g := range groups {
		gids[i] = strings.ReplaceAll(g.FullPath, "/", ":")
	}

	extra := userAttributeExtra(user, dormant)
	info := authentication.UserInfo{
		Username: user.Username,
		UID:      strconv.FormatInt(int64(user.ID), 10),
		Groups:   gids,
		Extra:    extra,
	}

	return info
}

func userAttributeGroups(user *gitlab.User, dormant bool) []string {
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
	if dormant {
		groups = append(groups, GroupDormant)
	}

	return groups
}

func userAttributeExtra(user *gitlab.User, dormant bool) map[string]authentication.ExtraValue {
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
	if dormant {
		attrs = append(attrs, AttributeDormant)
	}

	extra := map[string]authentication.ExtraValue{
		GitlabAttributesKey: attrs,
	}

	for _, attr := range user.CustomAttributes {
		extra[GitlabKeyNamespace+attr.Key] = []string{attr.Value}
	}

	return extra
}
