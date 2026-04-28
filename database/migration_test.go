package database

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Angus-Warman/gotrain/ext"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func openMemoryDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Discard,
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}

func openLoggedDB(t *testing.T, buf *bytes.Buffer) *gorm.DB {
	t.Helper()
	db := openMemoryDB(t)
	return db.Session(&gorm.Session{Logger: sqlLogger{buf}})
}

// columnNames returns the set of column names present in a table.
func columnNames(t *testing.T, db *gorm.DB, table string) map[string]bool {
	t.Helper()
	type col struct{ Name string }
	var cols []col
	if err := db.Raw("PRAGMA table_info(" + table + ")").Scan(&cols).Error; err != nil {
		t.Fatalf("PRAGMA table_info(%s): %v", table, err)
	}
	out := make(map[string]bool, len(cols))
	for _, c := range cols {
		out[c.Name] = true
	}
	return out
}

// tableExists reports whether a table is present in the SQLite schema.
func tableExists(t *testing.T, db *gorm.DB, table string) bool {
	t.Helper()
	var name string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
	return name == table
}

type simpleStruct struct {
	ID   uint
	Name string
}

type timestampStruct struct {
	ID        uint
	CreatedAt time.Time
}

func TestSemiAutoMigrate_CreatesTableWithPluralName(t *testing.T) {
	db := openMemoryDB(t)

	if err := semiAutoMigrate(db, map[string]any{"User": &simpleStruct{}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !tableExists(t, db, "Users") {
		t.Error("expected table 'Users' to exist")
	}
}

func TestSemiAutoMigrate_ColumnsMatchStructFields(t *testing.T) {
	db := openMemoryDB(t)

	if err := semiAutoMigrate(db, map[string]any{"User": &simpleStruct{}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cols := columnNames(t, db, "Users")
	for _, want := range []string{"id", "name"} {
		if !cols[want] {
			t.Errorf("expected column %q in Users, got: %v", want, cols)
		}
	}
}

func TestSemiAutoMigrate_MultipleModels(t *testing.T) {
	db := openMemoryDB(t)

	models := map[string]any{
		"User":    &simpleStruct{},
		"Product": &timestampStruct{},
	}
	if err := semiAutoMigrate(db, models); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, table := range []string{"Users", "Products"} {
		if !tableExists(t, db, table) {
			t.Errorf("expected table %q to exist", table)
		}
	}
}

func TestSemiAutoMigrate_EmptyModels(t *testing.T) {
	db := openMemoryDB(t)

	if err := semiAutoMigrate(db, map[string]any{}); err != nil {
		t.Fatalf("unexpected error for empty models: %v", err)
	}
}

func TestSemiAutoMigrate_IdempotentOnRerun(t *testing.T) {
	db := openMemoryDB(t)
	models := map[string]any{"User": &simpleStruct{}}

	if err := semiAutoMigrate(db, models); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if err := semiAutoMigrate(db, models); err != nil {
		t.Fatalf("second run (should be a no-op): %v", err)
	}
}

func TestSemiAutoMigrate_ReflectBuiltType(t *testing.T) {
	// Mirrors what buildInstances produces: a *T where T was assembled
	// via reflect.StructOf at runtime. This is the real production path.
	db := openMemoryDB(t)

	rt := reflect.StructOf([]reflect.StructField{
		{Name: "ID", Type: reflect.TypeFor[uint]()},
		{Name: "Score", Type: reflect.TypeFor[float64]()},
	})
	models := map[string]any{"Result": reflect.New(rt).Interface()}

	if err := semiAutoMigrate(db, models); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cols := columnNames(t, db, "Results")
	for _, want := range []string{"id", "score"} {
		if !cols[want] {
			t.Errorf("expected column %q in Results, got: %v", want, cols)
		}
	}
}

func TestSemiAutoMigrate_AddsNewColumn(t *testing.T) {
	db := openMemoryDB(t)

	// v1: initial schema
	v1 := reflect.StructOf([]reflect.StructField{
		{Name: "ID", Type: reflect.TypeFor[uint]()},
		{Name: "Name", Type: reflect.TypeFor[string]()},
	})

	if err := semiAutoMigrate(db, map[string]any{"User": reflect.New(v1).Interface()}); err != nil {
		t.Fatalf("v1 migration: %v", err)
	}

	// v2: Email column added
	v2 := reflect.StructOf([]reflect.StructField{
		{Name: "ID", Type: reflect.TypeFor[uint]()},
		{Name: "Name", Type: reflect.TypeFor[string]()},
		{Name: "Email", Type: reflect.TypeFor[string]()},
	})

	if err := semiAutoMigrate(db, map[string]any{"User": reflect.New(v2).Interface()}); err != nil {
		t.Fatalf("v2 migration: %v", err)
	}

	cols := columnNames(t, db, "Users")
	for _, want := range []string{"id", "name", "email"} {
		if !cols[want] {
			t.Errorf("expected column %q after schema evolution, got: %v", want, cols)
		}
	}
}
func TestSQL_CreateTable_ExactSQL(t *testing.T) {
	var buf bytes.Buffer
	db := openLoggedDB(t, &buf)

	rt := reflect.StructOf([]reflect.StructField{
		{Name: "ID", Type: reflect.TypeFor[uint]()},
		{Name: "Name", Type: reflect.TypeFor[string]()},
		{Name: "Email", Type: reflect.TypeFor[string]()},
	})
	if err := semiAutoMigrate(db, map[string]any{"User": reflect.New(rt).Interface()}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "CREATE TABLE `Users` (`id` integer PRIMARY KEY AUTOINCREMENT,`name` text,`email` text);\n"
	if buf.String() != want {
		t.Errorf("SQL mismatch\nwant: %q\n got: %q", want, buf.String())
	}
}

func TestSQL_AddColumn_ExactSQL(t *testing.T) {
	var buf bytes.Buffer
	db := openLoggedDB(t, &buf)

	v1 := reflect.StructOf([]reflect.StructField{
		{Name: "ID", Type: reflect.TypeFor[uint]()},
		{Name: "Name", Type: reflect.TypeFor[string]()},
	})

	if err := semiAutoMigrate(db, map[string]any{"User": reflect.New(v1).Interface()}); err != nil {
		t.Fatalf("v1 migration: %v", err)
	}

	buf.Reset()
	v2 := reflect.StructOf([]reflect.StructField{
		{Name: "ID", Type: reflect.TypeFor[uint]()},
		{Name: "Name", Type: reflect.TypeFor[string]()},
		{Name: "Email", Type: reflect.TypeFor[string]()},
	})

	if err := semiAutoMigrate(db, map[string]any{"User": reflect.New(v2).Interface()}); err != nil {
		t.Fatalf("v2 migration: %v", err)
	}

	want := "ALTER TABLE `Users` ADD `email` text;\n"
	if buf.String() != want {
		t.Errorf("SQL mismatch\nwant: %q\n got: %q", want, buf.String())
	}
}

func TestSQL_NoOutput_WhenSchemaUnchanged(t *testing.T) {
	var buf bytes.Buffer
	db := openLoggedDB(t, &buf)

	rt := reflect.StructOf([]reflect.StructField{
		{Name: "ID", Type: reflect.TypeFor[uint]()},
		{Name: "Name", Type: reflect.TypeFor[string]()},
	})
	model := map[string]any{"User": reflect.New(rt).Interface()}

	// First run — creates the table, produces SQL.
	if err := semiAutoMigrate(db, model); err != nil {
		t.Fatalf("first run: %v", err)
	}

	// Second run against the same DB — schema is identical, buffer must stay empty.
	buf.Reset()
	if err := semiAutoMigrate(db, model); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no SQL for unchanged schema, got:\n%s", buf.String())
	}
}

func TestSQL_SelectsNotCaptured(t *testing.T) {
	var buf bytes.Buffer
	db := openLoggedDB(t, &buf)

	rt := reflect.StructOf([]reflect.StructField{
		{Name: "ID", Type: reflect.TypeFor[uint]()},
	})

	if err := semiAutoMigrate(db, map[string]any{"User": reflect.New(rt).Interface()}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, line := range strings.Split(buf.String(), "\n") {
		if strings.HasPrefix(strings.TrimSpace(strings.ToUpper(line)), "SELECT") {
			t.Errorf("SELECT statement leaked into migration output: %s", line)
		}
	}
}

func TestSQL_ParseModels_MatchesStaticStruct(t *testing.T) {
	// Write a model.go to a temp dir — this is what ParseModels walks.
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.go")
	err := os.WriteFile(modelPath, []byte(`package model

type User struct {
	ID    uint
	Name  string
	Email string
}
`), 0644)

	// Equivalent as a real struct
	type User struct {
		ID    uint
		Name  string
		Email string
	}

	if err != nil {
		t.Fatalf("writing model.go: %v", err)
	}

	// Parse it the same way GenerateMigration does.
	models, err := ext.ParseModels(dir)
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}

	// Capture SQL from the parsed model.
	var parsedBuf bytes.Buffer
	parsedDB := openLoggedDB(t, &parsedBuf)
	if err := semiAutoMigrate(parsedDB, models); err != nil {
		t.Fatalf("semiAutoMigrate (parsed): %v", err)
	}

	// Capture SQL from the equivalent static struct.
	var staticBuf bytes.Buffer
	staticDB := openLoggedDB(t, &staticBuf)
	if err := semiAutoMigrate(staticDB, map[string]any{"User": &User{}}); err != nil {
		t.Fatalf("semiAutoMigrate (static): %v", err)
	}

	if parsedBuf.String() != staticBuf.String() {
		t.Errorf("SQL mismatch between parsed and static struct\nparsed: %q\nstatic: %q",
			parsedBuf.String(), staticBuf.String())
	}
}
