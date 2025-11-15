package models

import "time"

type PullRequest struct {
	PullRequestId     string     `json:"pull_request_id" db:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name" db:"pull_request_name"`
	AuthorId          string     `json:"author_id" db:"author_id"`
	Status            string     `json:"status" db:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers" db:"assigned_reviewers"`
	CreatedAt         *time.Time `db:"created_at"`
	MergedAt          *time.Time `db:"merged_at"`
}

type PullRequestShort struct {
	PullRequestId   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorId        string `json:"author_id"`
	Status          string `json:"status"`
}

type ReassignRequest struct {
	PullRequestId string `json:"pull_request_id"`
	OldUserId     string `json:"old_user_id"`
}
