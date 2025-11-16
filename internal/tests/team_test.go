package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/andro-kes/avito_test/internal/models"
)

func TestAddTeam(t *testing.T) {
	baseURL, _, _ := SetupTest(t)

	team := map[string]any{
		"team_name": "itest-team",
		"members": []map[string]any{
			{"user_id": "tu1", "username": "tu1", "is_active": true},
			{"user_id": "tu2", "username": "tu2", "is_active": true},
		},
	}

	b, err := json.Marshal(team)
	require.NoError(t, err)

	// Используем bytes.NewReader, потому что b это []byte
	req, err := http.NewRequestWithContext(context.Background(), "POST", baseURL+"/team/add/", bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)
}

func TestGetTeam(t *testing.T) {
	baseURL, _, _ := SetupTest(t)
	client := &http.Client{}

	// Создаём команду
	addBody := `{
		"team_name": "backend",
		"members": [
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true}
		]
	}`
	req, err := http.NewRequestWithContext(context.Background(), "POST", baseURL+"/team/add/", strings.NewReader(addBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)

	req1, err := http.NewRequestWithContext(context.Background(), "GET", baseURL+"/team/get?team_name=backend", http.NoBody)
	require.NoError(t, err)
	resp1, err := client.Do(req1)
	require.NoError(t, err)
	defer resp1.Body.Close()
	require.Equal(t, 200, resp1.StatusCode)

	var teamResp models.Team
	err = json.NewDecoder(resp1.Body).Decode(&teamResp)
	require.NoError(t, err)
	require.Equal(t, "backend", teamResp.TeamName)
	require.Len(t, teamResp.Members, 2)

	req2, err := http.NewRequestWithContext(context.Background(), "GET", baseURL+"/team/get/", http.NoBody)
	require.NoError(t, err)
	resp2, err := client.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, 404, resp2.StatusCode)

	req3, err := http.NewRequestWithContext(context.Background(), "GET", baseURL+"/team/get?team_name=nonexistent", http.NoBody)
	require.NoError(t, err)
	resp3, err := client.Do(req3)
	require.NoError(t, err)
	defer resp3.Body.Close()
	require.Equal(t, 404, resp3.StatusCode)
}
