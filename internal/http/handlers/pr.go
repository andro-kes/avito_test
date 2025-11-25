package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	prerrors "github.com/andro-kes/avito_test/internal/errors"
	logger "github.com/andro-kes/avito_test/internal/log"
	"github.com/andro-kes/avito_test/internal/models"
)

func (hm *HandlerManager) CreatePR(c *gin.Context) {
	var pr models.PullRequestShort
	if err := c.ShouldBindJSON(&pr); err != nil {
		logger.Log.Error("Не десериализовать объект", zap.Error(err))
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	ctx := c.Request.Context()
	exists, err := hm.PRService.CheckExistingPR(ctx, pr.PullRequestId)
	if err != nil {
		logger.Log.Error("Server error", zap.Error(err))
		c.AbortWithStatusJSON(500, prerrors.ErrServer)
		return
	}
	if exists {
		logger.Log.Error("PR существует", zap.Error(err))
		c.AbortWithStatusJSON(409, prerrors.ErrPRExists)
		return
	}

	createdPR, err := hm.PRService.CreatePR(ctx, &pr)
	if err != nil {
		logger.Log.Error("Server error", zap.Error(err))
		c.AbortWithStatusJSON(500, prerrors.ErrServer)
		return
	}

	c.JSON(201, gin.H{
		"pr": createdPR,
	})
}

func (hm *HandlerManager) MergePR(c *gin.Context) {
	var pr models.PullRequestShort
	if err := c.ShouldBindJSON(&pr); err != nil {
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	ctx := c.Request.Context()
	merged, err := hm.PRService.MergePR(ctx, pr.PullRequestId)
	if err != nil {
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	c.JSON(200, gin.H{
		"pr": merged,
	})
}

func (hm *HandlerManager) ReassignReviewer(c *gin.Context) {
	var r models.ReassignRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	ctx := c.Request.Context()
	err := hm.PRService.IsMerged(ctx, r.PullRequestId)
	if err != nil {
		if errors.Is(err, prerrors.ErrPRMerged) {
			c.AbortWithStatusJSON(409, prerrors.ErrPRMerged)
			return
		}
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	pr, replaced_by, err := hm.PRService.ReassignReviewer(ctx, r.PullRequestId, r.OldUserId)
	if err != nil {
		if errors.Is(err, prerrors.ErrNoCandidate) {
			c.AbortWithStatusJSON(409, prerrors.ErrNoCandidate)
			return
		}
		if errors.Is(err, prerrors.ErrNotAssigned) {
			c.AbortWithStatusJSON(409, prerrors.ErrNotAssigned)
			return
		}
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	c.JSON(200, gin.H{
		"pr":          pr,
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
		"user_id":       userId,
		"pull_requests": reviews,
	})
}
