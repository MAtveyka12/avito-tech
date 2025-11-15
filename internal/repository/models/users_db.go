package models

import (
	"time"

	"avito-tech/internal/domain/users"
)

type UserDB struct {
	UserID    string
	Username  string
	TeamName  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ToDomainUser(userDB *UserDB) *users.User {
	return &users.User{
		UserID:   userDB.UserID,
		Username: userDB.Username,
		TeamName: userDB.TeamName,
		IsActive: userDB.IsActive,
	}
}

func FromDomainUser(user *users.User) *UserDB {
	return &UserDB{
		UserID:   user.UserID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}
