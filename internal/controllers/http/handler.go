package http

import "avito-tech/internal/service"

type Handler struct {
	teamsService         service.TeamsService
	usersService         service.UsersService
	pullRequestsService  service.PullRequestsService
}

func NewHandler(
	teamsService service.TeamsService,
	usersService service.UsersService,
	pullRequestsService service.PullRequestsService,
) *Handler {
	return &Handler{
		teamsService:        teamsService,
		usersService:        usersService,
		pullRequestsService: pullRequestsService,
	}
}

