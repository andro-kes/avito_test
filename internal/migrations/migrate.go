package migrations

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	logger "github.com/andro-kes/avito_test/internal/log"
	"go.uber.org/zap"
)

//go:embed all:migrations
var migrationsFS embed.FS

type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if err := createMigrationsTable(ctx, pool); err != nil {
		return err
	}

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	applied, err := getAppliedMigrations(ctx, pool)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if applied[migration.Version] {
			logger.Log.Info("Migration already applied", 
				zap.Int("version", migration.Version),
				zap.String("name", migration.Name))
			continue
		}

		logger.Log.Info("Applying migration", 
			zap.Int("version", migration.Version),
			zap.String("name", migration.Name))

		if err := applyMigration(ctx, pool, migration); err != nil {
			return err
		}

		if err := markMigrationApplied(ctx, pool, migration.Version, migration.Name); err != nil {
			return err
		}

		logger.Log.Info("Migration applied successfully", 
			zap.Int("version", migration.Version))
	}

	return nil
}

func createMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	sql := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP NOT NULL DEFAULT NOW()
	)
	`
	_, err := pool.Exec(ctx, sql)
	return err
}

func loadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}

	migrationMap := make(map[int]*Migration)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Парсим имя файла: 000001_create_schema.up.sql или 000001_create_schema.down.sql
		parts := strings.Split(entry.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			continue
		}

		// Извлекаем имя миграции (без .up.sql или .down.sql)
		nameParts := strings.Split(entry.Name(), ".")
		if len(nameParts) < 3 {
			continue
		}
		name := strings.Join(nameParts[:len(nameParts)-2], ".")

		if migrationMap[version] == nil {
			migrationMap[version] = &Migration{
				Version: version,
				Name:    name,
			}
		}

		content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", entry.Name()))
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(entry.Name(), ".up.sql") {
			migrationMap[version].UpSQL = string(content)
		} else if strings.HasSuffix(entry.Name(), ".down.sql") {
			migrationMap[version].DownSQL = string(content)
		}
	}

	migrations := make([]Migration, 0, len(migrationMap))
	for _, m := range migrationMap {
		if m.UpSQL == "" {
			continue
		}
		migrations = append(migrations, *m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func getAppliedMigrations(ctx context.Context, pool *pgxpool.Pool) (map[int]bool, error) {
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, migration Migration) error {
	_, err := pool.Exec(ctx, migration.UpSQL)
	return err
}

func markMigrationApplied(ctx context.Context, pool *pgxpool.Pool, version int, name string) error {
	_, err := pool.Exec(ctx,
		"INSERT INTO schema_migrations (version, name) VALUES ($1, $2) ON CONFLICT (version) DO NOTHING",
		version, name)
	return err
}

