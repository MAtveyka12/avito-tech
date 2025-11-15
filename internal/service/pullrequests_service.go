package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"avito-tech/internal/domain/pullrequests"
	"avito-tech/internal/domain/teams"
	"avito-tech/internal/domain/users"
	"avito-tech/internal/repository/tx"
)

type PullRequestsService interface {
	CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*pullrequests.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (*pullrequests.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*pullrequests.PullRequest, string, error)
	GetPullRequestsByReviewer(ctx context.Context, reviewerID string) ([]*pullrequests.PullRequestShort, error)
	GetAllOpenPullRequests(ctx context.Context) ([]*pullrequests.PullRequest, error)
}

type pullRequestsService struct {
	prRepo    pullrequests.PullRequestsRepository
	usersRepo users.UsersRepository
	teamsRepo teams.TeamsRepository
	txManager tx.TxManager
}

func NewPullRequestsService(
	prRepo pullrequests.PullRequestsRepository,
	usersRepo users.UsersRepository,
	teamsRepo teams.TeamsRepository,
	txManager tx.TxManager,
) PullRequestsService {
	return &pullRequestsService{
		prRepo:    prRepo,
		usersRepo: usersRepo,
		teamsRepo: teamsRepo,
		txManager: txManager,
	}
}

func (s *pullRequestsService) CreatePullRequest(
	ctx context.Context, prID, prName, authorID string,
) (
	*pullrequests.PullRequest, error,
) {
	author, err := s.usersRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("author not found: %w", err)
	}

	candidates, err := s.teamsRepo.GetActiveMembersByTeamName(ctx, author.TeamName, authorID)
	if err != nil {
		return nil, fmt.Errorf("get team members: %w", err)
	}

	reviewers := selectRandomReviewers(candidates, 2)
	reviewerIDs := make([]string, 0, len(reviewers))

	for _, reviewer := range reviewers {
		reviewerIDs = append(reviewerIDs, reviewer.UserID)
	}

	now := time.Now()
	pr := &pullrequests.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            pullrequests.StatusOpen,
		AssignedReviewers: reviewerIDs,
		CreatedAt:         &now,
	}

	err = s.txManager.Do(ctx, func(txCtx context.Context) error {
		return s.prRepo.Create(txCtx, pr)
	})
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *pullRequestsService) MergePullRequest(ctx context.Context, prID string) (*pullrequests.PullRequest, error) {
	var pr *pullrequests.PullRequest

	var err error

	err = s.txManager.Do(ctx, func(txCtx context.Context) error {
		pr, err = s.prRepo.GetByID(txCtx, prID)
		if err != nil {
			return err
		}

		if pr.Status == pullrequests.StatusMerged {
			return nil
		}

		now := time.Now()

		pr.Status = pullrequests.StatusMerged
		pr.MergedAt = &now

		return s.prRepo.Update(txCtx, pr)
	})
	if err != nil {
		return nil, err
	}

	if pr.Status != pullrequests.StatusMerged {
		pr, err = s.prRepo.GetByID(ctx, prID)
		if err != nil {
			return nil, err
		}
	}

	return pr, nil
}

func (s *pullRequestsService) ReassignReviewer(
	ctx context.Context, prID, oldUserID string,
) (
	*pullrequests.PullRequest, string, error,
) {
	var pr *pullrequests.PullRequest

	var newReviewerID string

	var err error

	err = s.txManager.Do(ctx, func(txCtx context.Context) error {
		pr, err = s.prRepo.GetByID(txCtx, prID)
		if err != nil {
			return err
		}

		if pr.Status == pullrequests.StatusMerged {
			return pullrequests.ErrPullRequestMerged
		}

		found := false

		for _, reviewerID := range pr.AssignedReviewers {
			if reviewerID == oldUserID {
				found = true
				break
			}
		}

		if !found {
			return pullrequests.ErrNotAssigned
		}

		oldReviewer, err := s.usersRepo.GetByID(txCtx, oldUserID)
		if err != nil {
			return fmt.Errorf("old reviewer not found: %w", err)
		}

		candidates, err := s.teamsRepo.GetActiveMembersByTeamName(txCtx, oldReviewer.TeamName, oldUserID)
		if err != nil {
			return fmt.Errorf("get team members: %w", err)
		}

		availableCandidates := make([]*teams.TeamMember, 0)

		for _, candidate := range candidates {
			isAssigned := false

			for _, reviewerID := range pr.AssignedReviewers {
				if candidate.UserID == reviewerID {
					isAssigned = true
					break
				}
			}

			if !isAssigned {
				availableCandidates = append(availableCandidates, candidate)
			}
		}

		if len(availableCandidates) == 0 {
			return pullrequests.ErrNoCandidate
		}

		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		selected := availableCandidates[rng.Intn(len(availableCandidates))]

		newReviewerID = selected.UserID

		for i, reviewerID := range pr.AssignedReviewers {
			if reviewerID == oldUserID {
				pr.AssignedReviewers[i] = newReviewerID
				break
			}
		}

		return s.prRepo.Update(txCtx, pr)
	})
	if err != nil {
		return nil, "", err
	}

	return pr, newReviewerID, nil
}

func (s *pullRequestsService) GetPullRequestsByReviewer(
	ctx context.Context, reviewerID string,
) (
	[]*pullrequests.PullRequestShort, error,
) {
	prs, err := s.prRepo.GetByReviewerID(ctx, reviewerID)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (s *pullRequestsService) GetAllOpenPullRequests(ctx context.Context) ([]*pullrequests.PullRequest, error) {
	prs, err := s.prRepo.GetAllOpen(ctx)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func selectRandomReviewers(candidates []*teams.TeamMember, maxCount int) []*teams.TeamMember {
	if len(candidates) == 0 {
		return []*teams.TeamMember{}
	}

	count := maxCount
	if len(candidates) < count {
		count = len(candidates)
	}

	shuffled := make([]*teams.TeamMember, len(candidates))
	copy(shuffled, candidates)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}
