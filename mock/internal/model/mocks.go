package model

import (
	"strconv"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var (
	mockTokenAdmin    = "ADMIN-00000000000-TOKEN"
	mockTokenMock     = "MOCK-000000000000-TOKEN"
	mockTokenSecure   = "SECURE-0000000000-TOKEN"
	mockTokenPrivate  = "PRIVATE-000000000-TOKEN"
	mockTokenExternal = "EXTERNAL-00000000-TOKEN"
	mockTokenBot      = "BOT-0000000000000-TOKEN"
	mockTokenLocked   = "LOCKED-0000000000-TOKEN"
	mockTokenPristine = "PRISTINE-00000000-TOKEN"
	mockTokenDormant  = "DORMANT-000000000-TOKEN"
)

type Mocks struct {
	TokenPrefix string
	GroupCount  uint64
}

func (m *Mocks) Create(dao *DataAccess) error {
	if err := m.createTokens(dao.Tokens); err != nil {
		return err
	}

	if err := m.createUsers(dao.Users); err != nil {
		return err
	}

	if err := m.createGroups(dao.Groups); err != nil {
		return err
	}

	return nil
}

func (m *Mocks) createTokens(dao TokensAccess) error {
	models := map[string]int{
		mockTokenAdmin:    adminUser.ID,
		mockTokenMock:     mockUser.ID,
		mockTokenSecure:   secureUser.ID,
		mockTokenPrivate:  privateUser.ID,
		mockTokenExternal: externalUser.ID,
		mockTokenBot:      botUser.ID,
		mockTokenLocked:   lockedUser.ID,
		mockTokenPristine: pristineUser.ID,
		mockTokenDormant:  dormantUser.ID,
	}

	for t, u := range models {
		if err := dao.Create(m.TokenPrefix+t, u); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mocks) createUsers(dao UsersAccess) error {
	models := []*gitlab.User{
		&adminUser,
		&mockUser,
		&secureUser,
		&privateUser,
		&externalUser,
		&botUser,
		&lockedUser,
		&pristineUser,
		&dormantUser,
	}

	for _, u := range models {
		if err := dao.Create(u); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mocks) createGroups(dao GroupsAccess) error {
	if err := m.createGroupModels(dao); err != nil {
		return err
	}

	if err := m.createGroupAccociations(dao); err != nil {
		return err
	}

	if err := m.createGroupFeatures(dao); err != nil {
		return err
	}

	return nil
}

func (m *Mocks) createGroupModels(dao GroupsAccess) error {
	models := []*gitlab.Group{
		&coreGroup,
		&testGroup,
		&adminGroup,
		&userGroup,
		&specialGroup,
	}

	for _, g := range models {
		if err := dao.Create(g); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mocks) createGroupAccociations(dao GroupsAccess) error {
	assoc := map[int][]int{
		adminUser.ID: []int{
			coreGroup.ID, adminGroup.ID,
		},
		mockUser.ID: []int{
			coreGroup.ID, testGroup.ID, userGroup.ID,
		},
		secureUser.ID: []int{
			coreGroup.ID, testGroup.ID, userGroup.ID,
		},
		privateUser.ID: []int{
			coreGroup.ID, testGroup.ID, userGroup.ID,
		},
		externalUser.ID: []int{
			coreGroup.ID, userGroup.ID, specialGroup.ID,
		},
		botUser.ID: []int{
			coreGroup.ID, userGroup.ID, specialGroup.ID,
		},
		lockedUser.ID: []int{
			coreGroup.ID, userGroup.ID, specialGroup.ID,
		},
		pristineUser.ID: []int{
			coreGroup.ID, userGroup.ID, specialGroup.ID,
		},
		dormantUser.ID: []int{
			coreGroup.ID, userGroup.ID, specialGroup.ID,
		},
	}

	for u, gs := range assoc {
		for _, g := range gs {
			if err := dao.CreateUserAssociation(g, u); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Mocks) createGroupFeatures(dao GroupsAccess) error {
	baseID := testGroup.ID * 1000
	featureUsers := []int{
		mockUser.ID, secureUser.ID, privateUser.ID,
	}

	for i := uint64(1); i <= m.GroupCount; i++ {
		feat := strconv.FormatUint(i, 10)
		path := "green-" + feat
		name := "Green " + feat
		if i%2 == 0 {
			path = "blue-" + feat
			name = "Blue " + feat
		}

		group := &gitlab.Group{
			ID:          baseID + int(i),
			ParentID:    testGroup.ID,
			Description: "Feature selection pool",
			Name:        name,
			Path:        path,
			FullName:    "test / " + path,
			FullPath:    "test/" + path,
			CreatedAt:   &created,
			Visibility:  gitlab.PrivateVisibility,
		}

		if err := dao.Create(group); err != nil {
			return err
		}

		for _, u := range featureUsers {
			if err := dao.CreateUserAssociation(group.ID, u); err != nil {
				return err
			}
		}
	}

	return nil
}
