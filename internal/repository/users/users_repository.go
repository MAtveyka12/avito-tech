package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"avito-tech/internal/domain/users"
	"avito-tech/internal/repository/models"
	"avito-tech/internal/repository/tx"
)

type usersRepository struct {
	pool       *pgxpool.Pool
	ctxManager tx.CtxManager
}

func NewUsersRepository(pool *pgxpool.Pool, ctxManager tx.CtxManager) users.UsersRepository {
	return &usersRepository{
		pool:       pool,
		ctxManager: ctxManager,
	}
}

type pgxQuerier interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (r *usersRepository) getQuerier(ctx context.Context) pgxQuerier {
	if tx, err := r.ctxManager.FromContext(ctx); err == nil {
		return tx.Raw()
	}

	return r.pool
}

func (r *usersRepository) GetByID(ctx context.Context, userID string) (*users.User, error) {
	q := `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`

	var userDB models.UserDB

	err := r.getQuerier(ctx).QueryRow(ctx, q, userID).Scan(
		&userDB.UserID, &userDB.Username, &userDB.TeamName, &userDB.IsActive,
		&userDB.CreatedAt, &userDB.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, users.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return models.ToDomainUser(&userDB), nil
}

func (r *usersRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	q := `
		UPDATE users
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
		RETURNING username, team_name, created_at, updated_at
	`

	var userDB models.UserDB

	userDB.UserID = userID

	err := r.getQuerier(ctx).QueryRow(ctx, q, isActive, userID).Scan(
		&userDB.Username, &userDB.TeamName, &userDB.CreatedAt, &userDB.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users.ErrUserNotFound
		}

		return fmt.Errorf("set is active: %w", err)
	}

	userDB.IsActive = isActive

	return nil
}

func (r *usersRepository) GetActiveMembersByTeamName(
	ctx context.Context, teamName string, excludeUserID string,
) (
	[]*users.User, error,
) {
	q := `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE team_name = $1 AND is_active = true AND user_id != $2
		ORDER BY user_id
	`

	rows, err := r.getQuerier(ctx).Query(ctx, q, teamName, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("get active members: %w", err)
	}

	defer rows.Close()

	var result []*users.User

	for rows.Next() {
		var userDB models.UserDB
		if err := rows.Scan(
			&userDB.UserID, &userDB.Username, &userDB.TeamName, &userDB.IsActive,
			&userDB.CreatedAt, &userDB.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}

		result = append(result, models.ToDomainUser(&userDB))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return result, nil
}

func (r *usersRepository) BulkDeactivateByTeamName(ctx context.Context, teamName string) error {
	q := `
		UPDATE users
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE team_name = $1
	`

	_, err := r.getQuerier(ctx).Exec(ctx, q, teamName)
	if err != nil {
		return fmt.Errorf("bulk deactivate users: %w", err)
	}

	return nil
}
