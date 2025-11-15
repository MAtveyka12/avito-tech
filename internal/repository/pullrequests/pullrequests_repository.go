package pullrequests

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"avito-tech/internal/domain/pullrequests"
	"avito-tech/internal/repository/models"
	"avito-tech/internal/repository/tx"
)

type pullRequestsRepository struct {
	pool       *pgxpool.Pool
	ctxManager tx.CtxManager
}

func NewPullRequestsRepository(pool *pgxpool.Pool, ctxManager tx.CtxManager) pullrequests.PullRequestsRepository {
	return &pullRequestsRepository{
		pool:       pool,
		ctxManager: ctxManager,
	}
}

type pgxQuerier interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (r *pullRequestsRepository) getQuerier(ctx context.Context) pgxQuerier {
	if tx, err := r.ctxManager.FromContext(ctx); err == nil {
		return tx.Raw()
	}

	return r.pool
}

func (r *pullRequestsRepository) Create(ctx context.Context, pr *pullrequests.PullRequest) error {
	prDB := models.FromDomainPullRequest(pr)
	qPR := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, COALESCE($5, CURRENT_TIMESTAMP))
		RETURNING created_at, updated_at
	`

	var createdAt time.Time

	err := r.getQuerier(ctx).QueryRow(ctx, qPR,
		prDB.PullRequestID, prDB.PullRequestName, prDB.AuthorID, prDB.Status, prDB.CreatedAt,
	).Scan(&createdAt, &prDB.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return pullrequests.ErrPullRequestExists
		}

		return fmt.Errorf("create pull request: %w", err)
	}

	prDB.CreatedAt = createdAt

	for _, reviewerID := range pr.AssignedReviewers {
		qReviewer := `
			INSERT INTO pull_request_reviewers (pull_request_id, user_id)
			VALUES ($1, $2)
		`
		_, err := r.getQuerier(ctx).Exec(ctx, qReviewer, pr.PullRequestID, reviewerID)

		if err != nil {
			return fmt.Errorf("add reviewer: %w", err)
		}
	}

	return nil
}

func (r *pullRequestsRepository) GetByID(ctx context.Context, prID string) (*pullrequests.PullRequest, error) {
	qPR := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, updated_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	var prDB models.PullRequestDB

	err := r.getQuerier(ctx).QueryRow(ctx, qPR, prID).Scan(
		&prDB.PullRequestID, &prDB.PullRequestName, &prDB.AuthorID, &prDB.Status,
		&prDB.CreatedAt, &prDB.UpdatedAt, &prDB.MergedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pullrequests.ErrPullRequestNotFound
		}

		return nil, fmt.Errorf("get pull request: %w", err)
	}

	qReviewers := `
		SELECT user_id
		FROM pull_request_reviewers
		WHERE pull_request_id = $1
		ORDER BY user_id
	`

	rows, err := r.getQuerier(ctx).Query(ctx, qReviewers, prID)
	if err != nil {
		return nil, fmt.Errorf("get reviewers: %w", err)
	}

	defer rows.Close()

	var reviewers []string

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("scan reviewer: %w", err)
		}

		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviewers: %w", err)
	}

	return models.ToDomainPullRequest(&prDB, reviewers), nil
}

func (r *pullRequestsRepository) Update(ctx context.Context, pr *pullrequests.PullRequest) error {
	prDB := models.FromDomainPullRequest(pr)

	qPR := `
		UPDATE pull_requests
		SET pull_request_name = $1, status = $2, merged_at = $3, updated_at = CURRENT_TIMESTAMP
		WHERE pull_request_id = $4
		RETURNING created_at, updated_at
	`

	err := r.getQuerier(ctx).QueryRow(ctx, qPR,
		prDB.PullRequestName, prDB.Status, prDB.MergedAt, prDB.PullRequestID,
	).Scan(&prDB.CreatedAt, &prDB.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pullrequests.ErrPullRequestNotFound
		}

		return fmt.Errorf("update pull request: %w", err)
	}

	qDeleteReviewers := `DELETE FROM pull_request_reviewers WHERE pull_request_id = $1`

	_, err = r.getQuerier(ctx).Exec(ctx, qDeleteReviewers, pr.PullRequestID)
	if err != nil {
		return fmt.Errorf("delete reviewers: %w", err)
	}

	for _, reviewerID := range pr.AssignedReviewers {
		qReviewer := `
			INSERT INTO pull_request_reviewers (pull_request_id, user_id)
			VALUES ($1, $2)
		`
		_, err := r.getQuerier(ctx).Exec(ctx, qReviewer, pr.PullRequestID, reviewerID)

		if err != nil {
			return fmt.Errorf("add reviewer: %w", err)
		}
	}

	return nil
}

func (r *pullRequestsRepository) GetByReviewerID(
	ctx context.Context, reviewerID string,
) (
	[]*pullrequests.PullRequestShort, error,
) {
	q := `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		INNER JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.getQuerier(ctx).Query(ctx, q, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("get pull requests by reviewer: %w", err)
	}

	defer rows.Close()

	var result []*pullrequests.PullRequestShort

	for rows.Next() {
		var prDB models.PullRequestDB
		if err := rows.Scan(
			&prDB.PullRequestID, &prDB.PullRequestName, &prDB.AuthorID, &prDB.Status,
		); err != nil {
			return nil, fmt.Errorf("scan pull request: %w", err)
		}

		result = append(result, models.ToDomainPullRequestShort(&prDB))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pull requests: %w", err)
	}

	return result, nil
}

func (r *pullRequestsRepository) GetAllOpen(ctx context.Context) ([]*pullrequests.PullRequest, error) {
	q := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, updated_at, merged_at
		FROM pull_requests
		WHERE status = 'OPEN'
		ORDER BY created_at DESC
	`

	rows, err := r.getQuerier(ctx).Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("get all open pull requests: %w", err)
	}

	defer rows.Close()

	var result []*pullrequests.PullRequest

	for rows.Next() {
		var prDB models.PullRequestDB
		if err := rows.Scan(
			&prDB.PullRequestID, &prDB.PullRequestName, &prDB.AuthorID, &prDB.Status,
			&prDB.CreatedAt, &prDB.UpdatedAt, &prDB.MergedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pull request: %w", err)
		}

		qReviewers := `
			SELECT user_id
			FROM pull_request_reviewers
			WHERE pull_request_id = $1
			ORDER BY user_id
		`
		reviewerRows, err := r.getQuerier(ctx).Query(ctx, qReviewers, prDB.PullRequestID)
		if err != nil {
			return nil, fmt.Errorf("get reviewers: %w", err)
		}

		var reviewers []string

		for reviewerRows.Next() {
			var reviewerID string
			if err := reviewerRows.Scan(&reviewerID); err != nil {
				reviewerRows.Close()
				return nil, fmt.Errorf("scan reviewer: %w", err)
			}

			reviewers = append(reviewers, reviewerID)
		}

		reviewerRows.Close()

		if err := reviewerRows.Err(); err != nil {
			return nil, fmt.Errorf("iterate reviewers: %w", err)
		}

		result = append(result, models.ToDomainPullRequest(&prDB, reviewers))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pull requests: %w", err)
	}

	return result, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
