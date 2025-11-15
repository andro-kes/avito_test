package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	prerrors "github.com/andro-kes/avito_test/internal/errors"
	logger "github.com/andro-kes/avito_test/internal/log"
	"github.com/andro-kes/avito_test/internal/models"
	"github.com/andro-kes/avito_test/internal/repo"
	"github.com/andro-kes/avito_test/internal/repo/db"
)

type PRService struct {
	Repo repo.PRRepo
	Tx   db.Tx
}

func NewPRService(pool *pgxpool.Pool) *PRService {
	return &PRService{
		Repo: repo.NewPRRepo(pool),
		Tx:   db.NewTx(pool),
	}
}

func (ps *PRService) CreatePR(ctx context.Context, pr *models.PullRequestShort) (*models.PullRequest, error) {
	var pullRequest *models.PullRequest
	err := ps.Tx.RunInTx(ctx, func(ctx context.Context, q db.Querier) error {
		activeReviewers, err := ps.Repo.FindActiveReviewers(ctx, pr.AuthorId)
		if err != nil {
			return err
		}
		logger.Log.Info(fmt.Sprintf("Найдено %d кандидатов в ревьюеры", len(activeReviewers)))

		reviewers := random(activeReviewers)
		logger.Log.Info(
			fmt.Sprintf("Назначено %d ревьюера", len(reviewers)),
			zap.Any("reviewers", reviewers),
			zap.String("pr_id", pr.PullRequestId),
		)

		newPR, err := ps.Repo.CreatePR(ctx, q, pr, reviewers)
		if err != nil {
			return err
		}
		logger.Log.Info("PR создан")
		pullRequest = newPR
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pullRequest, nil
}

func random(r []string) []string {
	if len(r) <= 2 {
		return r
	}

	// Для выбора ревьюверов math/rand достаточно, не требуется криптографическая стойкость
	//nolint:gosec // G404: Use of weak random number generator is acceptable for non-security purposes
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rd.Shuffle(len(r), func(i, j int) {
		r[i], r[j] = r[j], r[i]
	})

	return r[:2]
}

func (ps *PRService) CheckExistingPR(ctx context.Context, id string) (bool, error) {
	return ps.Repo.CheckExistingPR(ctx, id)
}

func (ps *PRService) MergePR(ctx context.Context, id string) (*models.PullRequest, error) {
	return ps.Repo.MergePR(ctx, id)
}

func (ps *PRService) IsMerged(ctx context.Context, id string) error {
	return ps.Repo.IsMerged(ctx, id)
}

func (ps *PRService) ReassignReviewer(ctx context.Context, prId, oldUserId string) (*models.PullRequest, string, error) {
	var pr *models.PullRequest
	var replacedBy string
	err := ps.Tx.RunInTx(ctx, func(ctx context.Context, q db.Querier) error {
		replacement, err := ps.Repo.FindReplacementReviewers(ctx, prId, oldUserId)
		if err != nil {
			return err
		}

		if len(replacement) == 0 {
			return prerrors.ErrNoCandidate
		}

		replacedBy = random(replacement)[0]
		pr, err = ps.Repo.ReassignReviewer(ctx, q, prId, oldUserId, replacedBy)
		return err
	})

	return pr, replacedBy, err
}

func (us *PRService) GetReview(ctx context.Context, userId string) ([]models.PullRequestShort, error) {
	return us.Repo.GetReview(ctx, userId)
}
