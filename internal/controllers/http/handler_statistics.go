package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"avito-tech/internal/controllers/dto"
)

func (h *Handler) GetStatistics(c *gin.Context) {
	openPRs, err := h.pullRequestsService.GetAllOpenPullRequests(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})

		return
	}

	userAssignments := make(map[string]int)

	for _, pr := range openPRs {
		for _, reviewerID := range pr.AssignedReviewers {
			userAssignments[reviewerID]++
		}
	}

	prCount := len(openPRs)

	c.JSON(http.StatusOK, dto.StatisticsResponse{
		UserAssignments: userAssignments,
		PRCount:         prCount,
	})
}
