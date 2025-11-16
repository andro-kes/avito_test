package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

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

		reviewers := random(activeReviewers, 2)
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

func random(r []string, n int) []string {
	if len(r) <= 2 {
		return r
	}

	for i := len(r) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		j := int(jBig.Int64())
		r[i], r[j] = r[j], r[i]
	}

	if n > len(r) {
		return r
	}

	return r[:n]
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
		replacement, err := ps.Repo.FindReplacementReviewers(ctx, prId, []string{oldUserId})
		if err != nil {
			return err
		}

		if len(replacement) == 0 {
			return prerrors.ErrNoCandidate
		}

		replacedBy = random(replacement, 1)[0]
		pr, err = ps.Repo.ReassignReviewer(ctx, q, prId, oldUserId, replacedBy)
		return err
	})

	return pr, replacedBy, err
}

func (ps *PRService) GetReview(ctx context.Context, userId string) ([]models.PullRequestShort, error) {
	return ps.Repo.GetReview(ctx, userId)
}

func (ps *PRService) GetListByUsers(ctx context.Context, ids []string) (map[string]models.PullRequest, error) {
	prsMap := make(map[string]models.PullRequest)
	prs, err := ps.Repo.GetListByUsers(ctx, ids)
	if err != nil {
		return prsMap, err
	}

	for _, pr := range prs {
		prsMap[pr.PullRequestId] = pr
	}

	return prsMap, nil
}

func (ps *PRService) ReassignDeactivatedUsers(ctx context.Context, pr *models.PullRequest, ids []string) error {
	return ps.Tx.RunInTx(ctx, func(ctx context.Context, q db.Querier) error {
		replacement, err := ps.Repo.FindReplacementReviewers(ctx, pr.PullRequestId, ids)
		if err != nil {
			return err
		}

		if len(replacement) == 0 {
			return prerrors.ErrNoCandidate
		}

		replaced := random(replacement, len(replacement))

		deactivated := make(map[string]struct{}, len(ids))
		for _, u := range ids {
			deactivated[u] = struct{}{}
		}

		used := make(map[string]struct{}, 0)
		for _, u := range pr.AssignedReviewers {
			if _, ok := deactivated[u]; !ok {
				used[u] = struct{}{}
			}
		}
		used[pr.AuthorId] = struct{}{}

		repIdx := 0
		nextReplacement := func() (string, bool) {
			for repIdx < len(replaced) {
				c := replaced[repIdx]
				repIdx++
				if _, d := deactivated[c]; d {
					continue
				}
				if _, u := used[c]; u {
					continue
				}
				used[c] = struct{}{}
				return c, true
			}
			return "", false
		}

		newAssigned := make([]string, 0, len(pr.AssignedReviewers))
		for _, r := range pr.AssignedReviewers {
			if _, d := deactivated[r]; !d {
				newAssigned = append(newAssigned, r)
				continue
			}
			if cand, ok := nextReplacement(); ok {
				newAssigned = append(newAssigned, cand)
			}
		}

		return ps.Repo.ChangeDeactivatedReviewers(ctx, q, pr.PullRequestId, newAssigned)
	})
}
