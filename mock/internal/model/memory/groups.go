package memory

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock/internal/model"
)

type groupsAccess struct {
	s *storage
}

func (g *groupsAccess) Create(group *gitlab.Group) error {
	if _, ok := g.s.groups[group.ID]; ok {
		return model.ErrConflict
	}

	g.s.groups[group.ID] = group

	return nil
}

func (g *groupsAccess) CreateUserAssociation(gid, uid int) error {
	if _, ok := g.s.groups[gid]; !ok {
		return model.ErrNotFound
	}

	gids, ok := g.s.userGroups[uid]
	if !ok {
		g.s.userGroups[uid] = []int{gid}
		return nil
	}

	g.s.userGroups[uid] = append(gids, gid)

	return nil
}

func (g *groupsAccess) FindByUserIdentifier(uid, offset, limit int) ([]*gitlab.Group, error) {
	gids, ok := g.s.userGroups[uid]
	if !ok {
		return nil, model.ErrNotFound
	}

	high := offset + limit
	if offset < 0 || limit <= 0 {
		return []*gitlab.Group{}, nil
	}

	total := len(gids)
	if offset >= total {
		gids = nil
	} else if high >= total {
		gids = gids[offset:]
	} else {
		gids = gids[offset:high]
	}

	groups := make([]*gitlab.Group, len(gids))
	for i, gid := range gids {
		groups[i], _ = g.s.groups[gid]
	}

	return groups, nil
}

func (g *groupsAccess) CountByUserIdentifier(uid int) (int, error) {
	gids, ok := g.s.userGroups[uid]
	if !ok {
		return 0, model.ErrNotFound
	}

	return len(gids), nil
}
