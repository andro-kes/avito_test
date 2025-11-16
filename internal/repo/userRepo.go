package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	prerrors "github.com/andro-kes/avito_test/internal/errors"
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

func (ur *userRepo) CountReview(ctx context.Context, userId string) (int, error) {
	var exists bool
	err := ur.Pool.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)",
		userId,
	).Scan(&exists)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, prerrors.ErrNotFound
	}

	const sql = `
	SELECT COUNT(*) 
	FROM pull_requests 
	WHERE $1 = ANY(assigned_reviewers)
	`

	var cnt int
	err = ur.Pool.QueryRow(
		ctx,
		sql,
		userId,
	).Scan(&cnt)
	if err != nil {
		return 0, err
	}

	return cnt, nil
}

func (ur *userRepo) DeactivateUsers(ctx context.Context, q db.Querier, userIds []string) error {
	_, err := q.Exec(
		ctx,
		"UPDATE users SET is_active = false WHERE user_id = ANY($1)",
		userIds,
	)
	
	return err
}
