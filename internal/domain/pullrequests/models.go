package pullrequests

import "time"

type PullRequestStatus string

const (
	StatusOpen   PullRequestStatus = "OPEN"
	StatusMerged PullRequestStatus = "MERGED"
)

func (s PullRequestStatus) IsValid() bool {
	return s == StatusOpen || s == StatusMerged
}

type PullRequest struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          PullRequestStatus
	AssignedReviewers []string
	CreatedAt       *time.Time
	MergedAt        *time.Time
}

type PullRequestShort struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          PullRequestStatus
}

