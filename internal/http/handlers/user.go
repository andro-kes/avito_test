package handlers

import (
	"github.com/andro-kes/avito_test/internal/models"
	"github.com/gin-gonic/gin"
)

func (hm *HandlerManager) SetIsActive(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.AbortWithStatusJSON(404, gin.H{
			"code": "NOT_FOUND",
			"message": "user not found",
		})
		return 
	}

	ctx := c.Request.Context()
	if err := hm.UserService.SetIsActive(ctx, user.UserId, user.IsActive); err != nil {
		c.AbortWithStatusJSON(404, gin.H{
			"code": "NOT_FOUND",
			"message": "user not found",
		})
		return 
	}

	updatedUser, err := hm.UserService.GetUser(ctx, user.UserId)
	if err != nil {
		c.AbortWithStatusJSON(404, gin.H{
			"code": "NOT_FOUND",
			"message": "user not found",
		})
		return 
	}

	c.JSON(200, updatedUser)
}