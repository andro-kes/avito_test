package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/andro-kes/avito_test/internal/models"
	"github.com/stretchr/testify/require"
)

func TestCreatePR(t *testing.T) {
	baseURL, _, _ := SetupTest(t)
	

	// Создаем команду с минимум 2 активными пользователями (автор + ревьюер)
	team := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},   // автор
			{"user_id": "u2", "username": "Bob", "is_active": true},     // ревьюер
		},
	}
	body, _ := json.Marshal(team)
	resp, err := http.Post(baseURL+"/team/add/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	// Создаем PR от u1
	pr := map[string]any{
		"pull_request_id":   "pr-2001",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}
	body, _ = json.Marshal(pr)
	resp, err = http.Post(baseURL+"/pullRequest/create/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	var result map[string]models.PullRequestShort
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "pr-2001", result["pr"].PullRequestId)
}

func TestMergePR(t *testing.T) {
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
	resp, err := http.Post(baseURL+"/team/add/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	// Создаем PR
	pr := map[string]any{
		"pull_request_id":   "pr-2002",
		"pull_request_name": "Fix bug",
		"author_id":         "u1",
	}
	body, _ = json.Marshal(pr)
	resp, err = http.Post(baseURL+"/pullRequest/create/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	// Мержим PR
	mergeReq := map[string]any{
		"pull_request_id": "pr-2002",
	}
	body, _ = json.Marshal(mergeReq)
	resp, err = http.Post(baseURL+"/pullRequest/merge/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var mergeResp map[string]models.PullRequestShort
	err = json.NewDecoder(resp.Body).Decode(&mergeResp)
	require.NoError(t, err)
	require.Equal(t, "pr-2002", mergeResp["pr"].PullRequestId)
}

func TestReassignReviewer(t *testing.T) {
	baseURL, _, _ := SetupTest(t)

	// Создаем команду с минимум 4 пользователями
	team := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},   // автор
			{"user_id": "u2", "username": "Bob", "is_active": true},     // ревьюер
			{"user_id": "u3", "username": "Charlie", "is_active": true}, // ревьюер
			{"user_id": "u4", "username": "Dave", "is_active": true},    // резерв
		},
	}
	body, _ := json.Marshal(team)
	resp, err := http.Post(baseURL+"/team/add/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	// Создаем PR от u1 с **только u2** как ревьюером
	pr := map[string]any{
		"pull_request_id":   "pr-3001",
		"pull_request_name": "Improve code",
		"author_id":         "u1",
	}
	body, _ = json.Marshal(pr)
	resp, err = http.Post(baseURL+"/pullRequest/create/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	// Reassign u2 на u3 (резервного)
	reassign := map[string]any{
		"pull_request_id": "pr-3001",
		"old_user_id":     "u2",
	}
	body, _ = json.Marshal(reassign)
	resp, err = http.Post(baseURL+"/pullRequest/reassign/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	prData := result["pr"].(map[string]any)
	require.Equal(t, "pr-3001", prData["pull_request_id"])
	replacedBy := result["replaced_by"].(string)
	require.Equal(t, "u3", replacedBy)
}

func TestGetUserReview(t *testing.T) {
	baseURL, _, _ := SetupTest(t)
	

	// Создаем команду с автором и ревьюером
	team := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true}, // автор
			{"user_id": "u2", "username": "Bob", "is_active": true},   // ревьюер
		},
	}
	body, _ := json.Marshal(team)
	resp, err := http.Post(baseURL+"/team/add/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	// Создаем PR
	pr := map[string]any{
		"pull_request_id":   "pr-4001",
		"pull_request_name": "Add logging",
		"author_id":         "u1",
	}
	body, _ = json.Marshal(pr)
	resp, err = http.Post(baseURL+"/pullRequest/create/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	// Проверяем GetUserReview для u2
	resp, err = http.Get(baseURL + "/users/getReview/?user_id=u2")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, "u2", result["user_id"])
	reviews := result["pull_requests"].([]any)
	require.GreaterOrEqual(t, len(reviews), 1)
}
