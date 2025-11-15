package handlers

import (
	prerrors "github.com/andro-kes/avito_test/internal/errors"
	"github.com/andro-kes/avito_test/internal/models"
	"github.com/gin-gonic/gin"
)

func (hm *HandlerManager) AddTeam(c *gin.Context) {
	var team models.Team
	if err := c.ShouldBindBodyWithJSON(&team); err != nil {
		c.AbortWithStatusJSON(404, prerrors.ErrNotFound)
		return
	}

	ctx := c.Request.Context()
	if err := hm.TeamService.CheckUnique(ctx, team.TeamName); err != nil {
		c.AbortWithStatusJSON(400, prerrors.ErrTeamExists)
		return
	}

	if err := hm.TeamService.CreateTeamWithMembers(ctx, team); err != nil {
		c.AbortWithStatusJSON(500, prerrors.ErrServer)
		return
	}

	newTeam, err := hm.TeamService.GetTeam(ctx, team.TeamName)
	if err != nil {
		c.AbortWithStatusJSON(500, prerrors.ErrServer)
		return
	}

	c.JSON(201, newTeam)
}


func (hm *HandlerManager) GetTeam(c *gin.Context) {
	ctx := c.Request.Context()

	name, ok := c.GetQuery("team_name") 
	if !ok {
		c.AbortWithStatusJSON(400, prerrors.ErrNotFound)
		return 
	}

	team, err := hm.TeamService.GetTeam(ctx, name)
	if err != nil {
		c.AbortWithStatusJSON(400, prerrors.ErrNotFound)
		return 
	}

	c.JSON(200, team)
}