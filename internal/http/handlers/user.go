package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"

	prerrors "github.com/andro-kes/avito_test/internal/errors"
	"github.com/andro-kes/avito_test/internal/models"
)

func (hm *HandlerManager) SetIsActive(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.AbortWithStatusJSON(400, prerrors.ErrNotFound)
		return
	}

	ctx := c.Request.Context()
	if err := hm.UserService.SetIsActive(ctx, user.UserId, user.IsActive); err != nil {
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	updatedUser, err := hm.UserService.GetUser(ctx, user.UserId)
	if err != nil {
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	c.JSON(200, gin.H{
		"user": updatedUser,
	})
}

func (hm *HandlerManager) CountReview(c *gin.Context) {
	userId := c.Query("user_id")
	if userId == "" {
		c.AbortWithStatusJSON(400, prerrors.ErrNotFound)
		return
	}

	ctx := c.Request.Context()
	cnt, err := hm.UserService.CountReview(ctx, userId)
	if err != nil {
		if errors.Is(err, prerrors.ErrNotFound) {
			c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
			return
		}
		c.AbortWithStatusJSON(500, prerrors.ErrServer)
		return
	}

	c.JSON(200, gin.H{
		"user_id": userId,
		"reviews": cnt,
	})
}

type deactivatedRequest struct {
	UserIds []string `json:"user_ids"`
}

func (hm *HandlerManager) DeactivateUsers(c *gin.Context) {
	var ids deactivatedRequest
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.AbortWithStatusJSON(400, prerrors.ErrNotFound)
		return
	}

	if len(ids.UserIds) == 0 {
		c.AbortWithStatusJSON(400, prerrors.ErrNotFound)
		return
	}

	ctx := c.Request.Context()
	err := hm.UserService.DeactivateUsers(ctx, ids.UserIds)
	if err != nil {
		c.AbortWithStatusJSON(500, prerrors.ErrServer)
		return
	}

	prsMap, err := hm.PRService.GetListByUsers(ctx, ids.UserIds)
	if err != nil {
		c.AbortWithStatusJSON(500, prerrors.ErrServer)
		return
	}
	for _, pr := range prsMap {
		err := hm.PRService.ReassignDeactivatedUsers(ctx, &pr, ids.UserIds)
		if err != nil {
			c.AbortWithStatusJSON(500, prerrors.ErrServer)
			return
		}
	}

	c.JSON(200, "SUCCESS!")
}
