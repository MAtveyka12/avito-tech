package teams

import "context"

type TeamsRepository interface {
	Create(ctx context.Context, team *Team) error
	GetByName(ctx context.Context, name string) (*Team, error)
	GetActiveMembersByTeamName(ctx context.Context, teamName string, excludeUserID string) ([]*TeamMember, error)
}

