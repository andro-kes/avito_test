package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andro-kes/avito_test/internal/models"
	"github.com/andro-kes/avito_test/internal/repo"
	"github.com/andro-kes/avito_test/internal/repo/db"
)

type UserService struct {
	Repo repo.UserRepo
	Tx   db.Tx
}

func NewUserService(pool *pgxpool.Pool) *UserService {
	return &UserService{
		Repo: repo.NewUserRepo(pool),
		Tx:   db.NewTx(pool),
	}
}

func (us *UserService) SetIsActive(ctx context.Context, userId string, isActive bool) error {
	return us.Tx.RunInTx(ctx, func(ctx context.Context, q db.Querier) error {
		return us.Repo.SetIsActive(ctx, q, userId, isActive)
	})
}

func (us *UserService) GetUser(ctx context.Context, userId string) (*models.User, error) {
	return us.Repo.GetUser(ctx, userId)
}
