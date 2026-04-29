package parse

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

type FieldDef struct {
	Name string
	Type string // as a string, e.g. "string", "uint", "time.Time"
	Tag  string // raw struct tag, e.g. `gorm:"primaryKey"`
}

type StructDef struct {
	Name   string
	Fields []FieldDef
}

func ParseModels(appFolder string) (map[string]any, error) {
	structs, err := extractStructsFromFolder(appFolder)

	if err != nil {
		return nil, err
	}

	instances := buildInstances(structs)

	return instances, nil
}

func ParseModel(modelFilePath string) ([]FieldDef, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, modelFilePath, nil, 0)

	if err != nil {
		return nil, err
	}

	structs := extractStructs(file)

	if len(structs) != 1 {
		return nil, fmt.Errorf("Invalid number of structs in model file")
	}

	return structs[0].Fields, nil
}

func extractStructsFromFolder(appFolder string) ([]StructDef, error) {
	fset := token.NewFileSet()
	var allStructs []StructDef

	err := filepath.WalkDir(appFolder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "model.go" {
			return nil
		}

		slog.Debug("Extracting from: " + path)

		file, err := parser.ParseFile(fset, path, nil, 0)

		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		structs := extractStructs(file)

		allStructs = append(allStructs, structs...)

		return nil
	})

	return allStructs, err
}

func extractStructs(file *ast.File) []StructDef {
	var out []StructDef

	ast.Inspect(file, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)

		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)

		if !ok {
			return true
		}

		s := StructDef{Name: typeSpec.Name.Name}

		for _, field := range structType.Fields.List {

			typeName := exprToString(field.Type)

			tag := ""

			if field.Tag != nil {
				tag = strings.Trim(field.Tag.Value, "`")
			}

			if len(field.Names) == 0 {
				// Embedded field (e.g. gorm.Model)
				s.Fields = append(s.Fields, FieldDef{Name: "", Type: typeName, Tag: tag})
				continue
			}

			for _, name := range field.Names {
				s.Fields = append(s.Fields, FieldDef{Name: name.Name, Type: typeName, Tag: tag})
			}
		}

		out = append(out, s)

		return true
	})
	return out
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	default:
		return "interface{}"
	}
}

// typeMap covers the common cases. Unknown types fall back to string
// unless a gorm:"type:..." tag is present to carry the DDL hint.
var typeMap = map[string]reflect.Type{
	"string":     reflect.TypeFor[string](),
	"bool":       reflect.TypeFor[bool](),
	"int":        reflect.TypeFor[int](),
	"int8":       reflect.TypeFor[int8](),
	"int16":      reflect.TypeFor[int16](),
	"int32":      reflect.TypeFor[int32](),
	"int64":      reflect.TypeFor[int64](),
	"uint":       reflect.TypeFor[uint](),
	"uint8":      reflect.TypeFor[uint8](),
	"uint16":     reflect.TypeFor[uint16](),
	"uint32":     reflect.TypeFor[uint32](),
	"uint64":     reflect.TypeFor[uint64](),
	"float32":    reflect.TypeFor[float32](),
	"float64":    reflect.TypeFor[float64](),
	"time.Time":  reflect.TypeFor[time.Time](),
	"*time.Time": reflect.TypeFor[*time.Time](),
	"[]byte":     reflect.TypeFor[[]byte](),
}

func buildInstances(structs []StructDef) map[string]any {
	instances := make(map[string]any)
	for _, s := range structs {
		t := buildType(s)
		if t != nil {
			instances[s.Name] = reflect.New(t).Interface()
		}
	}
	return instances
}

func buildType(s StructDef) reflect.Type {
	var fields []reflect.StructField
	for _, f := range s.Fields {
		rt, ok := typeMap[f.Type]
		if !ok {
			// Unknown type: fall back to string.
			// If the field has a gorm:"type:..." tag, GORM will use that for DDL.
			rt = reflect.TypeFor[string]()
		}
		sf := reflect.StructField{
			Type: rt,
		}
		if f.Name == "" {
			// Embedded - mark it anonymous
			sf.Anonymous = true
			sf.Name = lastName(f.Type)
		} else {
			sf.Name = f.Name
		}
		if f.Tag != "" {
			sf.Tag = reflect.StructTag(f.Tag)
		}
		fields = append(fields, sf)
	}
	return reflect.StructOf(fields)
}

func lastName(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return s[i+1:]
		}
	}
	return s
}
