package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andro-kes/avito_test/internal/models"
	"github.com/andro-kes/avito_test/internal/repo/db"
)

type userRepo struct {
	Pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) UserRepo {
	return &userRepo{
		Pool: pool,
	}
}

func (ur *userRepo) UpsertUser(ctx context.Context, q db.Querier, name string, m models.TeamMember) error {
	sql := `
	INSERT INTO users (user_id, username, team_name, is_active)
	VALUES ($1,$2,$3,$4)
	ON CONFLICT (user_id) DO UPDATE
	SET username = EXCLUDED.username,
		team_name = EXCLUDED.team_name,
		is_active = EXCLUDED.is_active;
	`

	_, err := q.Exec(
		ctx,
		sql,
		m.UserID, m.Username, name, m.IsActive,
	)

	return err
}

func (ur *userRepo) SetIsActive(ctx context.Context, q db.Querier, userId string, isActive bool) error {
	sql := `
	UPDATE users
    SET is_active = $1
    WHERE user_id = $2
	`

	_, err := q.Exec(
		ctx,
		sql,
		isActive, userId,
	)

	return err
}

func (ur *userRepo) GetUser(ctx context.Context, userId string) (*models.User, error) {
	var user models.User
	err := ur.Pool.QueryRow(
		ctx,
		"SELECT user_id, username, team_name, is_active FROM users WHERE user_id=$1",
		userId,
	).Scan(&user.UserId, &user.Username, &user.TeamName, &user.IsActive)

	return &user, err
}
