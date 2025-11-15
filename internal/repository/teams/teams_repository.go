package teams

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"avito-tech/internal/domain/teams"
	"avito-tech/internal/repository/models"
	"avito-tech/internal/repository/tx"
)

type teamsRepository struct {
	pool       *pgxpool.Pool
	ctxManager tx.CtxManager
}

func NewTeamsRepository(pool *pgxpool.Pool, ctxManager tx.CtxManager) teams.TeamsRepository {
	return &teamsRepository{
		pool:       pool,
		ctxManager: ctxManager,
	}
}

type pgxQuerier interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (r *teamsRepository) getQuerier(ctx context.Context) pgxQuerier {
	if tx, err := r.ctxManager.FromContext(ctx); err == nil {
		return tx.Raw()
	}

	return r.pool
}

func (r *teamsRepository) Create(ctx context.Context, team *teams.Team) error {
	teamDB, membersDB := models.FromDomainTeam(team)

	qTeam := `
		INSERT INTO teams (team_name)
		VALUES ($1)
		RETURNING created_at, updated_at
	`

	err := r.getQuerier(ctx).QueryRow(ctx, qTeam, teamDB.Name).Scan(
		&teamDB.CreatedAt, &teamDB.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return teams.ErrTeamExists
		}

		return fmt.Errorf("create team: %w", err)
	}

	for _, member := range membersDB {
		qUser := `
			INSERT INTO users (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) 
			DO UPDATE SET 
				username = EXCLUDED.username,
				team_name = EXCLUDED.team_name,
				is_active = EXCLUDED.is_active,
				updated_at = CURRENT_TIMESTAMP
			RETURNING created_at, updated_at
		`

		err := r.getQuerier(ctx).QueryRow(ctx, qUser,
			member.UserID, member.Username, member.TeamName, member.IsActive,
		).Scan(&member.CreatedAt, &member.UpdatedAt)
		if err != nil {
			return fmt.Errorf("create/update user: %w", err)
		}
	}

	return nil
}

func (r *teamsRepository) GetByName(ctx context.Context, name string) (*teams.Team, error) {
	qTeam := `
		SELECT team_name, created_at, updated_at
		FROM teams
		WHERE team_name = $1
	`

	var teamDB models.TeamDB

	err := r.getQuerier(ctx).QueryRow(ctx, qTeam, name).Scan(
		&teamDB.Name, &teamDB.CreatedAt, &teamDB.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, teams.ErrTeamNotFound
		}

		return nil, fmt.Errorf("get team: %w", err)
	}

	qMembers := `
		SELECT user_id, username, is_active, created_at, updated_at
		FROM users
		WHERE team_name = $1
		ORDER BY user_id
	`

	rows, err := r.getQuerier(ctx).Query(ctx, qMembers, name)
	if err != nil {
		return nil, fmt.Errorf("get team members: %w", err)
	}

	defer rows.Close()

	var membersDB []*models.TeamMemberDB

	for rows.Next() {
		var member models.TeamMemberDB

		member.TeamName = name
		if err := rows.Scan(
			&member.UserID, &member.Username, &member.IsActive,
			&member.CreatedAt, &member.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}

		membersDB = append(membersDB, &member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate members: %w", err)
	}

	return models.ToDomainTeam(&teamDB, membersDB), nil
}

func (r *teamsRepository) GetActiveMembersByTeamName(
	ctx context.Context, teamName string, excludeUserID string,
) (
	[]*teams.TeamMember, error,
) {
	q := `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1 AND is_active = true AND user_id != $2
		ORDER BY user_id
	`

	rows, err := r.getQuerier(ctx).Query(ctx, q, teamName, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("get active members: %w", err)
	}

	defer rows.Close()

	var members []*teams.TeamMember

	for rows.Next() {
		var member teams.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}

		members = append(members, &member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate members: %w", err)
	}

	return members, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
