package users

import "context"

type UsersRepository interface {
	GetByID(ctx context.Context, userID string) (*User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) error
	GetActiveMembersByTeamName(ctx context.Context, teamName string, excludeUserID string) ([]*User, error)
	BulkDeactivateByTeamName(ctx context.Context, teamName string) error
}

