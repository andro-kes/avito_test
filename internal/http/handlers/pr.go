package handlers

import (
	prerrors "github.com/andro-kes/avito_test/internal/errors"
	"github.com/andro-kes/avito_test/internal/models"
	"github.com/gin-gonic/gin"
)

func (hm *HandlerManager) CreatePR(c *gin.Context) {
	var pr models.PullRequestShort
	if err := c.ShouldBindBodyWithJSON(&pr); err != nil {
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	ctx := c.Request.Context()
	err := hm.PRService.GetPR(ctx, pr.PullRequestId)
	if err != nil {
		c.AbortWithStatusJSON(409, prerrors.ErrPRExists)
		return
	}

	new, err := hm.PRService.CreatePR(ctx, &pr)
	if err != nil {
		c.AbortWithError(500, prerrors.ErrServer)
	}

	c.JSON(201, *new)
}


func (hm *HandlerManager) MergePR(c *gin.Context) {
	var pr models.PullRequestShort
	if err := c.ShouldBindBodyWithJSON(&pr); err != nil {
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return 
	}

	ctx := c.Request.Context()
	merged, err := hm.PRService.MergePR(ctx, pr.PullRequestId)
	if err != nil {
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
	}

	c.JSON(200, *merged)
}

type reassign struct {
	PullRequestId string `json:"pull_request_id"`
	OldUserId string `json:"old_reviewer_id"`
}
func (hm *HandlerManager) ReassignReviewer(c *gin.Context) {
	var r reassign
	if err := c.ShouldBindBodyWithJSON(&r); err != nil {
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	ctx := c.Request.Context()
	err := hm.PRService.IsMerged(ctx, r.PullRequestId)
	if err != nil {
		c.AbortWithStatusJSON(409, prerrors.ErrPRMerged)
		return
	}

	pr, replaced_by, err := hm.PRService.ReassignReviewer(ctx, r.PullRequestId, r.PullRequestId)
	if err != nil {
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	c.JSON(200, gin.H{
		"pr": pr,
		"replaced_by": replaced_by,
	})
}


func (hm *HandlerManager) GetUserReview(c *gin.Context) {
	userId := c.Query("user_id")

	ctx := c.Request.Context()
	reviews, err := hm.PRService.GetReview(ctx, userId)
	if err != nil {
		c.AbortWithStatusJSON(500, prerrors.ErrServer)
		return
	}

	c.JSON(200, gin.H{
		"user_id": userId,
		"pull_requests": reviews,
	})
}