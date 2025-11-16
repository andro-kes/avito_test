package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/andro-kes/avito_test/internal/models"
)

func TestCreatePR(t *testing.T) {
	baseURL, _, _ := SetupTest(t)

	// Создаем команду с минимум 2 активными пользователями
	team := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	body, _ := json.Marshal(team)
	req, err := http.NewRequestWithContext(context.Background(), "POST", baseURL+"/team/add/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)

	// Создаем PR от u1
	pr := map[string]any{
		"pull_request_id":   "pr-2001",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}
	body, _ = json.Marshal(pr)
	req, err = http.NewRequestWithContext(context.Background(), "POST", baseURL+"/pullRequest/create/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)

	var result map[string]models.PullRequestShort
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "pr-2001", result["pr"].PullRequestId)
}

func TestMergePR(t *testing.T) {
	baseURL, _, _ := SetupTest(t)
	client := &http.Client{}

	// Создаем команду
	team := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	body, _ := json.Marshal(team)
	req, err := http.NewRequestWithContext(context.Background(), "POST", baseURL+"/team/add/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)

	// Создаем PR
	pr := map[string]any{
		"pull_request_id":   "pr-2002",
		"pull_request_name": "Fix bug",
		"author_id":         "u1",
	}
	body, _ = json.Marshal(pr)
	req, err = http.NewRequestWithContext(context.Background(), "POST", baseURL+"/pullRequest/create/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)

	// Мержим PR
	mergeReq := map[string]any{"pull_request_id": "pr-2002"}
	body, _ = json.Marshal(mergeReq)
	req, err = http.NewRequestWithContext(context.Background(), "POST", baseURL+"/pullRequest/merge/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	var mergeResp map[string]models.PullRequestShort
	err = json.NewDecoder(resp.Body).Decode(&mergeResp)
	require.NoError(t, err)
	require.Equal(t, "pr-2002", mergeResp["pr"].PullRequestId)
}

func TestReassignReviewer(t *testing.T) {
	baseURL, _, _ := SetupTest(t)
	client := &http.Client{}

	// Создаем команду
	team := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
			{"user_id": "u4", "username": "Dave", "is_active": true},
		},
	}
	body, _ := json.Marshal(team)
	req, err := http.NewRequestWithContext(context.Background(), "POST", baseURL+"/team/add/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)

	// Создаем PR
	pr := map[string]any{
		"pull_request_id":   "pr-3001",
		"pull_request_name": "Improve code",
		"author_id":         "u1",
	}
	body, _ = json.Marshal(pr)
	req, err = http.NewRequestWithContext(context.Background(), "POST", baseURL+"/pullRequest/create/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)

	var prResp map[string]any
	err = json.NewDecoder(resp.Body).Decode(&prResp)
	require.NoError(t, err)

	// Берем назначенных ревьюеров из поля "pr" -> "assigned_reviewers"
	prData := prResp["pr"].(map[string]any)
	reviewersIface, ok := prData["assigned_reviewers"]
	require.True(t, ok, "PR должен содержать поле assigned_reviewers")
	reviewers := reviewersIface.([]any)
	require.GreaterOrEqual(t, len(reviewers), 1)
	oldUserID := reviewers[0].(string)

	// Reassign первого ревьюера
	reassign := map[string]any{
		"pull_request_id": "pr-3001",
		"old_user_id":     oldUserID,
	}
	body, _ = json.Marshal(reassign)
	req, err = http.NewRequestWithContext(context.Background(), "POST", baseURL+"/pullRequest/reassign/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	replacedBy := result["replaced_by"].(string)
	require.NotEqual(t, oldUserID, replacedBy)
}

func TestGetUserReview(t *testing.T) {
	baseURL, _, _ := SetupTest(t)
	client := &http.Client{}

	// Создаем команду
	team := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	body, _ := json.Marshal(team)
	req, err := http.NewRequestWithContext(context.Background(), "POST", baseURL+"/team/add/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)

	// Создаем PR
	pr := map[string]any{
		"pull_request_id":   "pr-4001",
		"pull_request_name": "Add logging",
		"author_id":         "u1",
	}
	body, _ = json.Marshal(pr)
	req, err = http.NewRequestWithContext(context.Background(), "POST", baseURL+"/pullRequest/create/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)

	// GET-запрос для получения ревью
	req, err = http.NewRequestWithContext(context.Background(), "GET", baseURL+"/users/getReview/?user_id=u2", http.NoBody)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "u2", result["user_id"])
	reviews := result["pull_requests"].([]any)
	require.GreaterOrEqual(t, len(reviews), 1)
}
