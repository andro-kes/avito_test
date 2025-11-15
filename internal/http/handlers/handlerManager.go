package handlers

import (
	"github.com/andro-kes/avito_test/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HandlerManager struct {
	TeamService *service.TeamService
	PRService *service.PRService
	UserService *service.UserService
}

func NewHandlerManager(pool *pgxpool.Pool) *HandlerManager {
	return &HandlerManager{
		TeamService: service.NewTeamService(pool),
		UserService: service.NewUserService(pool),
		PRService: service.NewPRService(pool),
	}
}