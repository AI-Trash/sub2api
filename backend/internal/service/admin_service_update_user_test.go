//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestAdminService_UpdateUser_AllowsPromotingUserToAdmin(t *testing.T) {
	baseRepo := &userRepoStub{
		user: &User{
			ID:          7,
			Email:       "user@example.com",
			Role:        RoleUser,
			Status:      StatusActive,
			Concurrency: 3,
		},
	}
	repo := &balanceUserRepoStub{userRepoStub: baseRepo}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		authCacheInvalidator: invalidator,
	}

	updated, err := svc.UpdateUser(context.Background(), 7, &UpdateUserInput{Role: RoleAdmin})

	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, RoleAdmin, updated.Role)
	require.Len(t, repo.updated, 1)
	require.Equal(t, RoleAdmin, repo.updated[0].Role)
	require.Equal(t, []int64{7}, invalidator.userIDs)
}

func TestAdminService_UpdateUser_RejectsDemotingLastActiveAdmin(t *testing.T) {
	baseRepo := &userRepoStub{
		user: &User{
			ID:          11,
			Email:       "admin@example.com",
			Role:        RoleAdmin,
			Status:      StatusActive,
			Concurrency: 5,
		},
		listPage: &pagination.PaginationResult{Total: 1, Page: 1, PageSize: 1, Pages: 1},
	}
	repo := &balanceUserRepoStub{userRepoStub: baseRepo}
	svc := &adminServiceImpl{userRepo: repo}

	updated, err := svc.UpdateUser(context.Background(), 11, &UpdateUserInput{Role: RoleUser})

	require.ErrorIs(t, err, ErrLastActiveAdminRoleChange)
	require.Nil(t, updated)
	require.Empty(t, repo.updated)
}

func TestAdminService_UpdateUser_AllowsDemotingAdminWhenAnotherActiveAdminExists(t *testing.T) {
	baseRepo := &userRepoStub{
		user: &User{
			ID:          12,
			Email:       "admin2@example.com",
			Role:        RoleAdmin,
			Status:      StatusActive,
			Concurrency: 5,
		},
		listPage: &pagination.PaginationResult{Total: 2, Page: 1, PageSize: 1, Pages: 2},
	}
	repo := &balanceUserRepoStub{userRepoStub: baseRepo}
	invalidator := &authCacheInvalidatorStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		authCacheInvalidator: invalidator,
	}

	updated, err := svc.UpdateUser(context.Background(), 12, &UpdateUserInput{Role: RoleUser})

	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, RoleUser, updated.Role)
	require.Len(t, repo.updated, 1)
	require.Equal(t, RoleUser, repo.updated[0].Role)
	require.Equal(t, []int64{12}, invalidator.userIDs)
}

func TestAdminService_UpdateUser_RejectsInvalidRole(t *testing.T) {
	svc := &adminServiceImpl{}

	updated, err := svc.UpdateUser(context.Background(), 1, &UpdateUserInput{Role: "owner"})

	require.ErrorIs(t, err, ErrInvalidUserRole)
	require.Nil(t, updated)
}
