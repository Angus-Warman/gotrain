package database

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Angus-Warman/gotrain/ext"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type sqlLogger struct{ w *bytes.Buffer }

func (l sqlLogger) Trace(_ context.Context, _ time.Time,
	fc func() (string, int64), _ error) {
	sql, _ := fc()
	slog.Debug(sql)

	if !strings.HasPrefix(sql, "SELECT") {
		fmt.Fprintf(l.w, "%s;\n", sql)
	}
}
func (l sqlLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface { return l }
func (l sqlLogger) Info(context.Context, string, ...any)             {}
func (l sqlLogger) Warn(context.Context, string, ...any)             {}
func (l sqlLogger) Error(context.Context, string, ...any)            {}

func GenerateMigration(appFolder string) error {
	slog.Debug("Generating migrations...")

	migrationsFolder := path.Join(appFolder, "migrations")

	err := setupMigrationsFolder(migrationsFolder)

	if err != nil {
		return err
	}

	// 1. Open in-memory DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormlogger.Discard})

	if err != nil {
		return err
	}

	// 2. Replay existing migrations in order
	err = RunMigrations(db, migrationsFolder)

	if err != nil {
		return err
	}

	// 3. Find and parse the models
	models, err := ext.ParseModels(appFolder)

	if err != nil {
		return err
	}

	// 4. Run AutoMigrate, capturing generated SQL into a buffer
	var buf bytes.Buffer
	db = db.Session(&gorm.Session{Logger: sqlLogger{&buf}})

	err = semiAutoMigrate(db, models)

	if err != nil {
		return err
	}

	if buf.Len() == 0 {
		slog.Info("No schema changes detected")
		return nil
	}

	// 5. Create a file
	words := strings.Fields(strings.ReplaceAll(buf.String(), "`", ""))
	if len(words) > 3 {
		words = words[:3]
	}
	slug := strings.Join(words, "_")

	fileName := fmt.Sprintf("%d_%s.sql", time.Now().UTC().Unix(), slug)

	filePath := path.Join(migrationsFolder, fileName)

	err = os.WriteFile(filePath, buf.Bytes(), 0644)

	if err != nil {
		return err
	}

	slog.Debug("Migration written to %s", "file", fileName)

	return err
}

func setupMigrationsFolder(migrationsFolder string) error {
	err := os.MkdirAll(migrationsFolder, 0755)

	if err != nil {
		return err
	}

	src := "./static/0_CREATE_TABLE___migrations.sql"
	dst := path.Join(migrationsFolder, "0_CREATE_TABLE___migrations.sql")

	return ext.Copy(src, dst)
}

func RunMigrations(db *gorm.DB, appPath string) error {
	migrationsPath := filepath.Join(appPath, "migrations")

	glob := path.Join(migrationsPath, "*.sql")

	existing, err := filepath.Glob(glob)

	if err != nil {
		return err
	}

	sort.Strings(existing)

	sqlDB, err := db.DB()

	if err != nil {
		return err
	}

	for _, f := range existing {
		sql, err := os.ReadFile(f)

		if err != nil {
			return err
		}

		_, err = sqlDB.Exec(string(sql))

		if err != nil {
			return err
		}

		slog.Info("Applied %s", "migration", f)
	}

	return nil
}

func semiAutoMigrate(db *gorm.DB, models map[string]any) error {
	for modelName, model := range models {
		tableName := modelName + "s"
		err := db.Table(tableName).AutoMigrate(model)

		if err != nil {
			return err
		}
	}

	return nil
}
