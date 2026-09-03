package codegen_test

import (
	"testing"

	"github.com/atombender/go-jsonschema/pkg/codegen"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

func TestPrimitiveTypeFromJSONSchemaTypeStringFormats(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		format       string
		extraImports bool
		expected     codegen.Type
	}{
		{
			name:         "date without extra imports",
			format:       "date",
			extraImports: false,
			expected:     codegen.PrimitiveType{Type: "string"},
		},
		{
			name:         "time without extra imports",
			format:       "time",
			extraImports: false,
			expected:     codegen.PrimitiveType{Type: "string"},
		},
		{
			name:         "date with extra imports",
			format:       "date",
			extraImports: true,
			expected:     namedType("types", "SerializableDate"),
		},
		{
			name:         "time with extra imports",
			format:       "time",
			extraImports: true,
			expected:     namedType("types", "SerializableTime"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := codegen.PrimitiveTypeFromJSONSchemaType(
				schemas.TypeNameString,
				tc.format,
				false,
				false,
				tc.extraImports,
				nil,
				nil,
				nil,
				nil,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertType(t, tc.expected, got)
		})
	}
}

func assertType(t *testing.T, expected, got codegen.Type) {
	t.Helper()

	expNamed, expIsNamed := expected.(codegen.NamedType)
	gotNamed, gotIsNamed := got.(codegen.NamedType)

	if expIsNamed || gotIsNamed {
		if !expIsNamed || !gotIsNamed {
			t.Fatalf("expected %T, got %T", expected, got)
		}

		if expNamed.Decl.Name != gotNamed.Decl.Name {
			t.Fatalf("expected named type %q, got %q", expNamed.Decl.Name, gotNamed.Decl.Name)
		}

		return
	}

	if expected != got {
		t.Fatalf("expected %#v, got %#v", expected, got)
	}
}

func namedType(pkgName, typeName string) codegen.NamedType {
	return codegen.NamedType{
		Package: &codegen.Package{
			QualifiedName: pkgName,
		},
		Decl: &codegen.TypeDecl{
			Name: typeName,
		},
	}
}
