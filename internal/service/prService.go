package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/andro-kes/avito_test/internal/models"
	"github.com/andro-kes/avito_test/internal/repo"
	"github.com/andro-kes/avito_test/internal/repo/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PRService struct {
	Repo repo.PRRepo
	Tx db.Tx
}

func NewPRService(pool *pgxpool.Pool) *PRService {
	return &PRService{
		Repo: repo.NewPRRepo(pool),
		Tx: db.NewTx(pool),
	}
}

func (ps *PRService) CreatePR(ctx context.Context, pr *models.PullRequestShort) (*models.PullRequest, error) {
	var pullRequest *models.PullRequest
	err := ps.Tx.RunInTx(ctx, func(ctx context.Context, q db.Querier) error {
		activeReviewers, err := ps.Repo.FindActiveReviewers(ctx, pr.AuthorId)
		if err != nil {
			return err
		}
		reviewers := random(activeReviewers)

		newPR, err := ps.Repo.CreatePR(ctx, q, pr, reviewers)
		if err != nil {
			return err
		}
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

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
    rd.Shuffle(len(r), func(i, j int) {
        r[i], r[j] = r[j], r[i]
    })

	return r[:2]
}

func (ps *PRService) GetPR(ctx context.Context, id string) error {
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
		replacement, err := ps.Repo.FindReplacementReviewers(ctx, prId)
		if err != nil {
			return err
		}
		// TODO check 0 candidates
		replacedBy = random(replacement)[0]
		pr, err = ps.Repo.ReassignReviewer(ctx, q, prId, oldUserId, replacedBy)
		return err
	})

	return pr, replacedBy, err
}

func (us *PRService) GetReview(ctx context.Context, userId string) ([]models.PullRequestShort, error) {
	return us.Repo.GetReview(ctx, userId)
}