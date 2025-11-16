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

	req, err := http.NewRequestWithContext(context.Background(), "POST", baseURL+"/team/add/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)

	setActiveBody := map[string]any{
		"user_id":   "u2",
		"is_active": false,
	}
	body, _ = json.Marshal(setActiveBody)

	req, err = http.NewRequestWithContext(context.Background(), "POST", baseURL+"/users/set_is_active/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	var result map[string]models.User
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, false, result["user"].IsActive)
}

func TestCountReview(t *testing.T) {
	baseURL, _, _ := SetupTest(t)
	client := &http.Client{}

	team := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
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

	pr := map[string]any{
		"pull_request_id":   "pr-1001",
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

	req, err = http.NewRequestWithContext(context.Background(), "GET", baseURL+"/users/countReview/?user_id=u3", http.NoBody)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	var countResult map[string]any
	err = json.NewDecoder(resp.Body).Decode(&countResult)
	require.NoError(t, err)
	require.Equal(t, "u3", countResult["user_id"])
	require.Equal(t, float64(1), countResult["reviews"])
}

func TestDeactivateUsers(t *testing.T) {
	baseURL, _, _ := SetupTest(t)
	client := &http.Client{}

	team := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	body, _ := json.Marshal(team)
	req, err := http.NewRequestWithContext(context.Background(), "POST", baseURL+"/team/add/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)

	deactivateBody := map[string]any{
		"user_ids": []string{"u3"},
	}
	body, _ = json.Marshal(deactivateBody)
	req, err = http.NewRequestWithContext(context.Background(), "POST", baseURL+"/users/deactivate/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	var result map[string][]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	require.Equal(t, []string{"u3"}, result["deactivated"])
}
