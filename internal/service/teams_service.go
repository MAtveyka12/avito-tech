package service

import (
	"context"

	"avito-tech/internal/domain/teams"
	"avito-tech/internal/repository/tx"
)

type TeamsService interface {
	CreateTeam(ctx context.Context, team *teams.Team) (*teams.Team, error)
	GetTeam(ctx context.Context, teamName string) (*teams.Team, error)
}

type teamsService struct {
	repo      teams.TeamsRepository
	txManager tx.TxManager
}

func NewTeamsService(repo teams.TeamsRepository, txManager tx.TxManager) TeamsService {
	return &teamsService{
		repo:      repo,
		txManager: txManager,
	}
}

func (s *teamsService) CreateTeam(ctx context.Context, team *teams.Team) (*teams.Team, error) {
	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		return s.repo.Create(txCtx, team)
	})
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (s *teamsService) GetTeam(ctx context.Context, teamName string) (*teams.Team, error) {
	team, err := s.repo.GetByName(ctx, teamName)
	if err != nil {
		return nil, err
	}

	return team, nil
}

