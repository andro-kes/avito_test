package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/andro-kes/avito_test/internal/models"
	"github.com/stretchr/testify/require"
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

	b, _ := json.Marshal(team)
	resp, err := http.Post(baseURL+"/team/add/", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)
}

func TestGetTeam(t *testing.T) {
	baseURL, _, _ := SetupTest(t)

	addBody := `{
		"team_name": "backend",
		"members": [
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true}
		]
	}`
	resp, err := http.Post(fmt.Sprintf("%s/team/add/", baseURL), "application/json", strings.NewReader(addBody))
	require.NoError(t, err)
	require.Equal(t, 201, resp.StatusCode)

	// Тест успешного получения команды
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/team/get/?team_name=backend", baseURL), nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var teamResp models.Team
	err = json.NewDecoder(resp.Body).Decode(&teamResp)
	require.NoError(t, err)
	require.Equal(t, "backend", teamResp.TeamName)
	require.Len(t, teamResp.Members, 2)

	// Тест запроса без параметра
	req2, _ := http.NewRequest("GET", fmt.Sprintf("%s/team/get/", baseURL), nil)
	resp2, err := client.Do(req2)
	require.NoError(t, err)
	require.Equal(t, 400, resp2.StatusCode)

	// Тест запроса несуществующей команды
	req3, _ := http.NewRequest("GET", fmt.Sprintf("%s/team/get/?team_name=unknown", baseURL), nil)
	resp3, err := client.Do(req3)
	require.NoError(t, err)
	require.Equal(t, 400, resp3.StatusCode)
}
