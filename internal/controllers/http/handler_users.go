package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"avito-tech/internal/controllers/dto"
	"avito-tech/internal/domain/users"
)

func (h *Handler) SetIsActive(c *gin.Context) {
	var req dto.SetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})

		return
	}

	user, err := h.usersService.SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		if err == users.ErrUserNotFound {
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

	c.JSON(http.StatusOK, gin.H{
		"user": dto.ToUserResponse(user),
	})
}

func (h *Handler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "user_id is required",
			},
		})

		return
	}

	prs, err := h.pullRequestsService.GetPullRequestsByReviewer(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})

		return
	}

	prResponses := make([]dto.PullRequestShortResponse, 0, len(prs))
	for _, pr := range prs {
		prResponses = append(prResponses, dto.ToPullRequestShortResponse(pr))
	}

	c.JSON(http.StatusOK, dto.GetReviewResponse{
		UserID:       userID,
		PullRequests: prResponses,
	})
}

func (h *Handler) BulkDeactivate(c *gin.Context) {
	var req dto.BulkDeactivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})

		return
	}

	err := h.usersService.BulkDeactivateByTeamName(c.Request.Context(), req.TeamName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Users deactivated successfully",
	})
}
