package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	prerrors "github.com/andro-kes/avito_test/internal/errors"
	"github.com/andro-kes/avito_test/internal/models"
	"github.com/andro-kes/avito_test/internal/repo/db"
)

type prRepo struct {
	Pool *pgxpool.Pool
}

func NewPRRepo(pool *pgxpool.Pool) PRRepo {
	return &prRepo{
		Pool: pool,
	}
}

func (p *prRepo) CreatePR(ctx context.Context, q db.Querier, pr *models.PullRequestShort, reviewers []string) (*models.PullRequest, error) {
	status := "OPEN"
	if pr.Status != "" {
		status = pr.Status
	}

	sql := `
	INSERT INTO pull_requests 
    (pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at)
	VALUES 
    ($1, $2, $3, $4, $5, $6, $7)
	RETURNING 
    pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at
	`

	var pullRequest models.PullRequest
	err := q.QueryRow(
		ctx,
		sql,
		pr.PullRequestId, pr.PullRequestName, pr.AuthorId, status, reviewers, time.Now(), nil,
	).Scan(
		&pullRequest.PullRequestId, &pullRequest.PullRequestName,
		&pullRequest.AuthorId, &pullRequest.Status,
		&pullRequest.AssignedReviewers, &pullRequest.CreatedAt, &pullRequest.MergedAt,
	)

	return &pullRequest, err
}

func (p *prRepo) FindActiveReviewers(ctx context.Context, authorId string) ([]string, error) {
	sql := `
        SELECT user_id
        FROM users
        WHERE team_name = (
            SELECT team_name
            FROM users
            WHERE user_id = $1
        )
        AND is_active = TRUE
        AND user_id <> $1
    `

	rows, err := p.Pool.Query(ctx, sql, authorId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var active []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		active = append(active, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return active, nil
}

func (p *prRepo) CheckExistingPR(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := p.Pool.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)",
		id,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (p *prRepo) MergePR(ctx context.Context, id string) (*models.PullRequest, error) {
	const sql = `
	UPDATE pull_requests
	SET
		status = 'MERGED',
		merged_at = COALESCE(merged_at, NOW())
	WHERE pull_request_id = $1
	RETURNING
		pull_request_id,
		pull_request_name,
		author_id,
		status,
		assigned_reviewers,
		merged_at
	`

	var pr models.PullRequest
	err := p.Pool.QueryRow(
		ctx,
		sql,
		id,
	).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&pr.Status,
		&pr.AssignedReviewers,
		&pr.MergedAt,
	)

	return &pr, err
}

func (p *prRepo) IsMerged(ctx context.Context, id string) error {
	var isMerged bool
	err := p.Pool.QueryRow(
		ctx,
		"SELECT merged_at IS NOT NULL FROM pull_requests WHERE pull_request_id = $1",
		id,
	).Scan(&isMerged)

	if err != nil {
		return err
	}

	if isMerged {
		return prerrors.ErrPRMerged
	}

	return nil
}

func (p *prRepo) FindReplacementReviewers(ctx context.Context, prID, oldUserId string) ([]string, error) {
	sql := `
	SELECT u.user_id
	FROM users u
	INNER JOIN users old_reviewer ON u.team_name = old_reviewer.team_name
	INNER JOIN pull_requests pr ON pr.pull_request_id = $1
	WHERE old_reviewer.user_id = $2
	AND u.is_active = TRUE
	AND u.user_id <> pr.author_id
	AND u.user_id <> $2
	AND u.user_id <> ALL(pr.assigned_reviewers)
    `

	rows, err := p.Pool.Query(ctx, sql, prID, oldUserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replace []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		replace = append(replace, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return replace, nil
}

func (p *prRepo) ReassignReviewer(ctx context.Context, q db.Querier, prId, oldUserId, replacedBy string) (*models.PullRequest, error) {
	var isAssigned bool
	err := q.QueryRow(
		ctx,
		"SELECT $1 = ANY(assigned_reviewers) FROM pull_requests WHERE pull_request_id = $2",
		oldUserId, prId,
	).Scan(&isAssigned)
	if err != nil {
		return nil, prerrors.ErrNotFound
	}
	if !isAssigned {
		return nil, prerrors.ErrNotAssigned
	}

	const sql = `
	UPDATE pull_requests
	SET assigned_reviewers = array_replace(assigned_reviewers, $1, $2)
	WHERE pull_request_id = $3
	AND $1 = ANY(assigned_reviewers)
	AND status = 'OPEN'
	RETURNING pull_request_id, pull_request_name, author_id, status, assigned_reviewers
	`

	var pr models.PullRequest
	err = q.QueryRow(
		ctx,
		sql,
		oldUserId, replacedBy, prId,
	).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&pr.Status,
		&pr.AssignedReviewers,
	)

	if err != nil {
		return nil, err
	}

	return &pr, nil
}

func (p *prRepo) GetReview(ctx context.Context, userId string) ([]models.PullRequestShort, error) {
	const sql = `
	SELECT 
	pull_request_id, pull_request_name, author_id, status
	FROM pull_requests
	WHERE $1 = ANY(assigned_reviewers)
	`

	prs := make([]models.PullRequestShort, 0, 4)
	rows, err := p.Pool.Query(
		ctx,
		sql,
		userId,
	)
	if err != nil {
		return prs, err
	}
	defer rows.Close()

	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(
			&pr.PullRequestId,
			&pr.PullRequestName,
			&pr.AuthorId,
			&pr.Status,
		); err != nil {
			return prs, err
		}
		prs = append(prs, pr)
	}

	return prs, nil
}
