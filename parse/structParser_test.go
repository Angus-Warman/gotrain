package parse

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTemp writes content to a temp file and returns its path.
// The file is removed when the test ends.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "model*.go")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestParseModel_BasicFields(t *testing.T) {
	src := `package model

type User struct {
	ID    uint
	Name  string
	Email string
	Age   int
}
`
	path := writeTemp(t, src)
	fields, err := ParseModel(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []FieldDef{
		{Name: "ID", Type: "uint", Tag: ""},
		{Name: "Name", Type: "string", Tag: ""},
		{Name: "Email", Type: "string", Tag: ""},
		{Name: "Age", Type: "int", Tag: ""},
	}
	assertFields(t, fields, want)
}

func TestParseModel_StructTags(t *testing.T) {
	src := `package model

type Product struct {
	ID    uint   ` + "`" + `gorm:"primaryKey"` + "`" + `
	Name  string ` + "`" + `gorm:"not null" json:"name"` + "`" + `
}
`
	path := writeTemp(t, src)
	fields, err := ParseModel(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []FieldDef{
		{Name: "ID", Type: "uint", Tag: `gorm:"primaryKey"`},
		{Name: "Name", Type: "string", Tag: `gorm:"not null" json:"name"`},
	}
	assertFields(t, fields, want)
}

func TestParseModel_EmbeddedField(t *testing.T) {
	src := `package model

type Order struct {
	gorm.Model
	Total float64
}
`
	path := writeTemp(t, src)
	fields, err := ParseModel(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Embedded gorm.Model should appear as Name="" Type="gorm.Model"
	if len(fields) != 2 {
		t.Fatalf("want 2 fields, got %d", len(fields))
	}
	if fields[0].Name != "" {
		t.Errorf("embedded field Name: want %q, got %q", "", fields[0].Name)
	}
	if fields[0].Type != "gorm.Model" {
		t.Errorf("embedded field Type: want %q, got %q", "gorm.Model", fields[0].Type)
	}
	if fields[1].Name != "Total" || fields[1].Type != "float64" {
		t.Errorf("second field: want {Total float64}, got {%s %s}", fields[1].Name, fields[1].Type)
	}
}

func TestParseModel_PointerAndSliceTypes(t *testing.T) {
	src := `package model

type Article struct {
	CreatedAt *time.Time
	Tags      []string
}
`
	path := writeTemp(t, src)
	fields, err := ParseModel(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []FieldDef{
		{Name: "CreatedAt", Type: "*time.Time"},
		{Name: "Tags", Type: "[]string"},
	}
	assertFields(t, fields, want)
}

func TestParseModel_SelectorType(t *testing.T) {
	src := `package model

type Event struct {
	HappenedAt time.Time
}
`
	path := writeTemp(t, src)
	fields, err := ParseModel(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []FieldDef{
		{Name: "HappenedAt", Type: "time.Time"},
	}
	assertFields(t, fields, want)
}

func TestParseModel_MultipleFieldsSameType(t *testing.T) {
	// Go allows "First, Last string" on one line — the parser expands these.
	src := `package model

type Person struct {
	First, Last string
}
`
	path := writeTemp(t, src)
	fields, err := ParseModel(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []FieldDef{
		{Name: "First", Type: "string"},
		{Name: "Last", Type: "string"},
	}
	assertFields(t, fields, want)
}

func TestParseModel_ErrorOnZeroStructs(t *testing.T) {
	src := `package model

// no structs here
var Foo = 1
`
	path := writeTemp(t, src)
	_, err := ParseModel(path)
	if err == nil {
		t.Fatal("expected error for file with no structs, got nil")
	}
}

func TestParseModel_ErrorOnMultipleStructs(t *testing.T) {
	src := `package model

type Foo struct{ X string }
type Bar struct{ Y int }
`
	path := writeTemp(t, src)
	_, err := ParseModel(path)
	if err == nil {
		t.Fatal("expected error for file with multiple structs, got nil")
	}
}

func TestParseModel_ErrorOnInvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.go")
	_, err := ParseModel(path)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestParseModel_ErrorOnSyntaxError(t *testing.T) {
	src := `package model

type Broken struct {
	// missing closing brace
`
	path := writeTemp(t, src)
	_, err := ParseModel(path)
	if err == nil {
		t.Fatal("expected parse error for invalid Go syntax, got nil")
	}
}

// assertFields checks that got matches want exactly (order-sensitive).
func assertFields(t *testing.T, got, want []FieldDef) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("field count: want %d, got %d\n  want: %+v\n  got:  %+v", len(want), len(got), want, got)
	}
	for i, w := range want {
		g := got[i]
		if g.Name != w.Name {
			t.Errorf("field[%d].Name: want %q, got %q", i, w.Name, g.Name)
		}
		if g.Type != w.Type {
			t.Errorf("field[%d].Type: want %q, got %q", i, w.Type, g.Type)
		}
		if g.Tag != w.Tag {
			t.Errorf("field[%d].Tag: want %q, got %q", i, w.Tag, g.Tag)
		}
	}
}
