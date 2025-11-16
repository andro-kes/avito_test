package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/andro-kes/avito_test/internal/models"
	"github.com/stretchr/testify/require"
)

func TestSetIsActive(t *testing.T) {
	baseURL, _, _ := SetupTest(t)

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

	setActiveBody := map[string]any{
		"user_id":  "u2",
		"is_active": false,
	}
	body, _ = json.Marshal(setActiveBody)
	resp, err = http.Post(baseURL+"/users/set_is_active/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]models.User
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, false, result["user"].IsActive)
}

func TestCountReview(t *testing.T) {
	baseURL, _, _ := SetupTest(t)

	team := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	body, _ := json.Marshal(team)
	resp, err := http.Post(baseURL+"/team/add/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	pr := map[string]any{
		"pull_request_id":   "pr-1001",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}
	body, _ = json.Marshal(pr)
	resp, err = http.Post(baseURL+"/pullRequest/create/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	resp, err = http.Get(baseURL + "/users/countReview/?user_id=u3")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var countResult map[string]any
	err = json.NewDecoder(resp.Body).Decode(&countResult)
	require.NoError(t, err)
	require.Equal(t, "u3", countResult["user_id"])
	require.Equal(t, float64(1), countResult["reviews"])
}

func TestDeactivateUsers(t *testing.T) {
	baseURL, _, _ := SetupTest(t)

	team := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	body, _ := json.Marshal(team)
	resp, err := http.Post(baseURL+"/team/add/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	deactivateBody := map[string]any{
		"user_ids": []string{"u3"},
	}
	body, _ = json.Marshal(deactivateBody)
	resp, err = http.Post(baseURL+"/users/deactivate/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var deactivateResp string
	err = json.NewDecoder(resp.Body).Decode(&deactivateResp)
	require.NoError(t, err)
	require.Equal(t, "SUCCESS!", deactivateResp)
}
