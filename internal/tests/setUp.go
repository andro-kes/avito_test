package tests

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/andro-kes/avito_test/internal/http/handlers"
	logger "github.com/andro-kes/avito_test/internal/log"
	"github.com/andro-kes/avito_test/internal/migrations"
)

func SetupTest(t *testing.T) (baseURL string, db *pgxpool.Pool, router *gin.Engine) {
	t.Helper()

	logger.Init()

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	opts := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15",
		Env: []string{
			"POSTGRES_USER=pgtest",
			"POSTGRES_PASSWORD=pgtest",
			"POSTGRES_DB=pgtest",
		},
	}

	resource, err := pool.RunWithOptions(opts, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(t, err)
	resource.Expire(900)

	hostPort := resource.GetHostPort("5432/tcp")
	dsn := fmt.Sprintf("postgres://pgtest:pgtest@%s/pgtest?sslmode=disable", hostPort)

	var dbPool *pgxpool.Pool
	require.NoError(t, pool.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var err error
		dbPool, err = pgxpool.New(ctx, dsn)
		if err != nil {
			return err
		}
		return dbPool.Ping(ctx)
	}))
	db = dbPool

	require.NoError(t, migrations.ApplyMigrations(context.Background(), db))

	gin.SetMode(gin.TestMode)
	router = gin.Default()

	hm := handlers.NewHandlerManager(db)

	team := router.Group("/team/")
	team.POST("add/", hm.AddTeam)
	team.GET("get/", hm.GetTeam)

	user := router.Group("/users/")
	user.POST("set_is_active/", hm.SetIsActive)
	user.GET("getReview/", hm.GetUserReview)
	user.GET("countReview/", hm.CountReview)
	user.POST("deactivate/", hm.DeactivateUsers)

	pr := router.Group("/pullRequest/")
	pr.POST("create/", hm.CreatePR)
	pr.POST("merge/", hm.MergePR)
	pr.POST("reassign/", hm.ReassignReviewer)

	ts := httptest.NewServer(router)

	t.Cleanup(func() {
        ts.Close()
        if db != nil { db.Close() }
        _ = pool.Purge(resource)
    })

	return ts.URL, db, router
}