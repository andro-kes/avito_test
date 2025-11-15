package repo

import (
	"context"

	"github.com/andro-kes/avito_test/internal/models"
	"github.com/andro-kes/avito_test/internal/repo/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type teamRepo struct {
	Pool *pgxpool.Pool
}

func NewTeamRepo(pool *pgxpool.Pool) TeamRepo {
	return &teamRepo{
		Pool: pool,
	}
}

func (tr *teamRepo) CheckUnique(ctx context.Context, name string) error {
	var team models.Team
	err := tr.Pool.QueryRow(
		ctx,
		"SELECT team_name FROM teams WHERE team_name=$1",
		name,
	).Scan(&team.TeamName)

	return err
}

func (tr *teamRepo) CreateTeam(ctx context.Context, q db.Querier, name string) error {
    _, err := q.Exec(
        ctx,
        "INSERT INTO teams (team_name) VALUES ($1)",
        name,
    )
    return err
}

func (tr *teamRepo) GetTeam(ctx context.Context, name string) (*models.Team, error) {
	var team models.Team
	err := tr.Pool.QueryRow(
		ctx, 
		"SELECT team_name, members FROM teams WHERE team_name=$1",
		name,
	).Scan(&team.TeamName, &team.Members)

	return &team, err
}