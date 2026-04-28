package cmd

import (
	"strings"
	"testing"
)

func TestModelPropertiesFromStrings(t *testing.T) {
	tests := []struct {
		input []string
		want  []ModelProperty
	}{
		{
			input: []string{"name"},
			want: []ModelProperty{
				{Name: "Name", Type: "string", Tag: ""},
			},
		},
		{
			input: []string{"name", "age:int", "score:float", "active:bool"},
			want: []ModelProperty{
				{Name: "Name", Type: "string", Tag: ""},
				{Name: "Age", Type: "int", Tag: ""},
				{Name: "Score", Type: "float32", Tag: ""},
				{Name: "Active", Type: "string", Tag: ""},
			},
		},
		{
			// Type aliases all resolve correctly
			input: []string{"views:integer", "price:real", "rating:number"},
			want: []ModelProperty{
				{Name: "Views", Type: "int", Tag: ""},
				{Name: "Price", Type: "float32", Tag: ""},
				{Name: "Rating", Type: "float32", Tag: ""},
			},
		},
		{
			// Name is capitalised regardless of input casing
			input: []string{"firstName", "last_name", "EMAIL"},
			want: []ModelProperty{
				{Name: "FirstName", Type: "string", Tag: ""},
				{Name: "Last_name", Type: "string", Tag: ""},
				{Name: "EMAIL", Type: "string", Tag: ""},
			},
		},
		{
			input: []string{"email:required"},
			want: []ModelProperty{
				{Name: "Email", Type: "string", Tag: "`gorm:\"not null\"`"},
			},
		},
		{
			input: []string{"slug:unique"},
			want: []ModelProperty{
				{Name: "Slug", Type: "string", Tag: "`gorm:\"unique\"`"},
			},
		},
		{
			input: []string{"email:required:unique"},
			want: []ModelProperty{
				{Name: "Email", Type: "string", Tag: "`gorm:\"not null;unique\"`"},
			},
		},
		{
			input: []string{"quantity:int:required", "price:float:required:unique"},
			want: []ModelProperty{
				{Name: "Quantity", Type: "int", Tag: "`gorm:\"not null\"`"},
				{Name: "Price", Type: "float32", Tag: "`gorm:\"not null;unique\"`"},
			},
		},
		{
			// Unknown modifiers are silently ignored
			input: []string{"title:foobar"},
			want: []ModelProperty{
				{Name: "Title", Type: "string", Tag: ""},
			},
		},
		{
			// Empty slice returns empty slice, no error
			input: []string{},
			want:  []ModelProperty{},
		},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.input, " "), func(t *testing.T) {
			got, err := modelPropertiesFromStrings(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("length mismatch: want %d, got %d\nwant: %+v\ngot:  %+v",
					len(tt.want), len(got), tt.want, got)
			}
			for i, w := range tt.want {
				g := got[i]
				if g.Name != w.Name {
					t.Errorf("[%d] Name: want %q, got %q", i, w.Name, g.Name)
				}
				if g.Type != w.Type {
					t.Errorf("[%d] Type: want %q, got %q", i, w.Type, g.Type)
				}
				if g.Tag != w.Tag {
					t.Errorf("[%d] Tag: want %q, got %q", i, w.Tag, g.Tag)
				}
			}
		})
	}
}
