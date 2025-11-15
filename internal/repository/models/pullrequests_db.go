package models

import (
	"time"

	"avito-tech/internal/domain/pullrequests"
)

type PullRequestDB struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	MergedAt        *time.Time
}

type PullRequestReviewerDB struct {
	PullRequestID string
	UserID        string
	CreatedAt     time.Time
}

func ToDomainPullRequest(prDB *PullRequestDB, reviewers []string) *pullrequests.PullRequest {
	status := pullrequests.PullRequestStatus(prDB.Status)
	if !status.IsValid() {
		status = pullrequests.StatusOpen
	}

	var mergedAt *time.Time
	if prDB.MergedAt != nil && !prDB.MergedAt.IsZero() {
		mergedAt = prDB.MergedAt
	}

	var createdAt *time.Time
	if !prDB.CreatedAt.IsZero() {
		createdAt = &prDB.CreatedAt
	}

	return &pullrequests.PullRequest{
		PullRequestID:     prDB.PullRequestID,
		PullRequestName:   prDB.PullRequestName,
		AuthorID:          prDB.AuthorID,
		Status:            status,
		AssignedReviewers: reviewers,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}
}

func ToDomainPullRequestShort(prDB *PullRequestDB) *pullrequests.PullRequestShort {
	status := pullrequests.PullRequestStatus(prDB.Status)
	if !status.IsValid() {
		status = pullrequests.StatusOpen
	}

	return &pullrequests.PullRequestShort{
		PullRequestID:   prDB.PullRequestID,
		PullRequestName: prDB.PullRequestName,
		AuthorID:        prDB.AuthorID,
		Status:          status,
	}
}

func FromDomainPullRequest(pr *pullrequests.PullRequest) *PullRequestDB {
	var mergedAt *time.Time
	if pr.MergedAt != nil {
		mergedAt = pr.MergedAt
	}

	var createdAt time.Time
	if pr.CreatedAt != nil {
		createdAt = *pr.CreatedAt
	}

	return &PullRequestDB{
		PullRequestID:   pr.PullRequestID,
		PullRequestName: pr.PullRequestName,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
		MergedAt:        mergedAt,
		CreatedAt:       createdAt,
	}
}
