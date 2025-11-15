package models

import (
	"time"

	"avito-tech/internal/domain/teams"
)

type TeamDB struct {
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TeamMemberDB struct {
	TeamName  string
	UserID    string
	Username  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ToDomainTeam(teamDB *TeamDB, membersDB []*TeamMemberDB) *teams.Team {
	members := make([]*teams.TeamMember, 0, len(membersDB))
	for _, m := range membersDB {
		members = append(members, &teams.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	return &teams.Team{
		Name:    teamDB.Name,
		Members: members,
	}
}

func FromDomainTeam(team *teams.Team) (*TeamDB, []*TeamMemberDB) {
	teamDB := &TeamDB{
		Name: team.Name,
	}

	membersDB := make([]*TeamMemberDB, 0, len(team.Members))
	for _, m := range team.Members {
		membersDB = append(membersDB, &TeamMemberDB{
			TeamName: team.Name,
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	return teamDB, membersDB
}
