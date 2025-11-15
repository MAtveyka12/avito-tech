package service

import (
	"context"

	"avito-tech/internal/domain/users"
	"avito-tech/internal/repository/tx"
)

type UsersService interface {
	GetUser(ctx context.Context, userID string) (*users.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*users.User, error)
	BulkDeactivateByTeamName(ctx context.Context, teamName string) error
}

type usersService struct {
	repo      users.UsersRepository
	txManager tx.TxManager
}

func NewUsersService(repo users.UsersRepository, txManager tx.TxManager) UsersService {
	return &usersService{
		repo:      repo,
		txManager: txManager,
	}
}

func (s *usersService) GetUser(ctx context.Context, userID string) (*users.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *usersService) SetIsActive(ctx context.Context, userID string, isActive bool) (*users.User, error) {
	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		return s.repo.SetIsActive(txCtx, userID, isActive)
	})
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *usersService) BulkDeactivateByTeamName(ctx context.Context, teamName string) error {
	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		return s.repo.BulkDeactivateByTeamName(txCtx, teamName)
	})
}

