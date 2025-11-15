package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"avito-tech/internal/controllers/dto"
	"avito-tech/internal/domain/pullrequests"
)

func (h *Handler) CreatePullRequest(c *gin.Context) {
	var req dto.CreatePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})

		return
	}

	pr, err := h.pullRequestsService.CreatePullRequest(
		c.Request.Context(),
		req.PullRequestID,
		req.PullRequestName,
		req.AuthorID,
	)
	if err != nil {
		if err == pullrequests.ErrPullRequestExists {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "PR_EXISTS",
					Message: "PR id already exists",
				},
			})

			return
		}

		if err.Error() == "author not found: user not found" {
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

	c.JSON(http.StatusCreated, gin.H{
		"pr": dto.ToPullRequestResponse(pr),
	})
}

func (h *Handler) MergePullRequest(c *gin.Context) {
	var req dto.MergePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})

		return
	}

	pr, err := h.pullRequestsService.MergePullRequest(c.Request.Context(), req.PullRequestID)
	if err != nil {
		if err == pullrequests.ErrPullRequestNotFound {
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
		"pr": dto.ToPullRequestResponse(pr),
	})
}

func (h *Handler) ReassignReviewer(c *gin.Context) {
	var req dto.ReassignReviewerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})

		return
	}

	pr, newReviewerID, err := h.pullRequestsService.ReassignReviewer(
		c.Request.Context(),
		req.PullRequestID,
		req.OldUserID,
	)
	if err != nil {
		if err == pullrequests.ErrPullRequestNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "NOT_FOUND",
					Message: "resource not found",
				},
			})

			return
		}

		if err == pullrequests.ErrPullRequestMerged {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "PR_MERGED",
					Message: "cannot reassign on merged PR",
				},
			})

			return
		}

		if err == pullrequests.ErrNotAssigned {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "NOT_ASSIGNED",
					Message: "reviewer is not assigned to this PR",
				},
			})

			return
		}

		if err == pullrequests.ErrNoCandidate {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "NO_CANDIDATE",
					Message: "no active replacement candidate in team",
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

	c.JSON(http.StatusOK, dto.ReassignResponse{
		PR:         dto.ToPullRequestResponse(pr),
		ReplacedBy: newReviewerID,
	})
}
