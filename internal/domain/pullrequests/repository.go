package pullrequests

import "context"

type PullRequestsRepository interface {
	Create(ctx context.Context, pr *PullRequest) error
	GetByID(ctx context.Context, prID string) (*PullRequest, error)
	Update(ctx context.Context, pr *PullRequest) error
	GetByReviewerID(ctx context.Context, reviewerID string) ([]*PullRequestShort, error)
	GetAllOpen(ctx context.Context) ([]*PullRequest, error)
}

