package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	prerrors "github.com/andro-kes/avito_test/internal/errors"
	"github.com/andro-kes/avito_test/internal/models"
	"github.com/andro-kes/avito_test/internal/repo/db"
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
	var exists bool
	err := tr.Pool.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)",
		name,
	).Scan(&exists)
	if err != nil {
		return prerrors.ErrServer
	}

	if exists {
		return prerrors.ErrTeamExists
	}

	return nil
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
	var teamName string
	err := tr.Pool.QueryRow(
		ctx,
		"SELECT team_name FROM teams WHERE team_name = $1",
		name,
	).Scan(&teamName)
	if err != nil {
		return nil, prerrors.ErrNotFound
	}

	rows, err := tr.Pool.Query(
		ctx,
		"SELECT user_id, username, is_active FROM users WHERE team_name = $1 ORDER BY user_id",
		name,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]models.TeamMember, 0)
	for rows.Next() {
		var member models.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &models.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}
