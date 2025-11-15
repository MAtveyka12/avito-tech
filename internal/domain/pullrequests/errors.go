package pullrequests

import "errors"

var (
	ErrPullRequestNotFound = errors.New("pull request not found")
	ErrPullRequestExists   = errors.New("pull request already exists")
	ErrPullRequestMerged    = errors.New("pull request is merged")
	ErrNotAssigned         = errors.New("reviewer is not assigned")
	ErrNoCandidate         = errors.New("no active replacement candidate")
)

