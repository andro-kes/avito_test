package repo

import (
	"context"

	"github.com/andro-kes/avito_test/internal/models"
	"github.com/andro-kes/avito_test/internal/repo/db"
)

type PRRepo interface {
	CreatePR(ctx context.Context, q db.Querier, pr *models.PullRequestShort, reviewers []string) (*models.PullRequest, error)
	FindActiveReviewers(ctx context.Context, author_id string) ([]string, error)
	CheckExistingPR(ctx context.Context, id string) (bool, error)
	MergePR(ctx context.Context, id string) (*models.PullRequest, error)
	IsMerged(ctx context.Context, id string) error
	FindReplacementReviewers(ctx context.Context, prID, oldUserId string) ([]string, error)
	GetReview(ctx context.Context, userId string) ([]models.PullRequestShort, error)
	ReassignReviewer(ctx context.Context, q db.Querier, prId, oldUserId, replacedBy string) (*models.PullRequest, error)
}

type TeamRepo interface {
	CheckUnique(ctx context.Context, name string) error
	CreateTeam(ctx context.Context, q db.Querier, name string) error
	GetTeam(ctx context.Context, name string) (*models.Team, error)
}

type UserRepo interface {
	GetUser(ctx context.Context, userId string) (*models.User, error)
	SetIsActive(ctx context.Context, q db.Querier, userId string, isActive bool) error
	CountReview(ctx context.Context, userId string) (int, error)
	UpsertUser(ctx context.Context, q db.Querier, name string, m models.TeamMember) error
}
