package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andro-kes/avito_test/internal/models"
	"github.com/andro-kes/avito_test/internal/repo"
	"github.com/andro-kes/avito_test/internal/repo/db"
)

type TeamService struct {
	TeamRepo repo.TeamRepo
	UserRepo repo.UserRepo
	Tx       db.Tx
}

func NewTeamService(pool *pgxpool.Pool) *TeamService {
	return &TeamService{
		TeamRepo: repo.NewTeamRepo(pool),
		UserRepo: repo.NewUserRepo(pool),
		Tx:       db.NewTx(pool),
	}
}

func (ts *TeamService) CheckUnique(ctx context.Context, name string) error {
	return ts.TeamRepo.CheckUnique(ctx, name)
}

func (ts *TeamService) CreateTeamWithMembers(ctx context.Context, team models.Team) error {
	return ts.Tx.RunInTx(ctx, func(ctx context.Context, q db.Querier) error {
		err := ts.TeamRepo.CreateTeam(ctx, q, team.TeamName)
		if err != nil {
			return err
		}
		for _, m := range team.Members {
			err = ts.UserRepo.UpsertUser(ctx, q, team.TeamName, m)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (ts *TeamService) GetTeam(ctx context.Context, name string) (*models.Team, error) {
	return ts.TeamRepo.GetTeam(ctx, name)
}
