package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"avito-tech/internal/controllers/dto"
	"avito-tech/internal/domain/teams"
)

func (h *Handler) CreateTeam(c *gin.Context) {
	var req dto.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})

		return
	}

	teamMembers := make([]*teams.TeamMember, 0, len(req.Members))
	for _, m := range req.Members {
		teamMembers = append(teamMembers, &teams.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	team := &teams.Team{
		Name:    req.TeamName,
		Members: teamMembers,
	}

	createdTeam, err := h.teamsService.CreateTeam(c.Request.Context(), team)
	if err != nil {
		if err == teams.ErrTeamExists {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "TEAM_EXISTS",
					Message: "team_name already exists",
				},
			})

			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"team": dto.ToTeamResponse(createdTeam),
	})
}

func (h *Handler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "team_name is required",
			},
		})

		return
	}

	team, err := h.teamsService.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		if err == teams.ErrTeamNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "NOT_FOUND",
					Message: "resource not found",
				},
			})

			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})

		return
	}

	c.JSON(http.StatusOK, dto.ToTeamResponse(team))
}
