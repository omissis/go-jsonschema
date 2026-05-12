package tests_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/atombender/go-jsonschema/pkg/generator"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

var (
	errExit *exec.ExitError

	basicConfig = generator.Config{
		SchemaMappings:     []generator.SchemaMapping{},
		ExtraImports:       true,
		DefaultPackageName: "github.com/example/test",
		DefaultOutputName:  "-",
		ResolveExtensions:  []string{".json", ".yaml"},
		YAMLExtensions:     []string{".yaml", ".yml"},
		Warner: func(message string) {
			log.Printf("[from warner] %s", message)
		},
		Tags: []string{"json", "yaml", "mapstructure"},
	}
)

// func TestDebug(t *testing.T) {
// 	t.Parallel()

// 	testExampleFile(t, basicConfig, "./data/core/some/file.json")
// }

func TestCore(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/core")
}

func TestOmitBoth(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.Tags = []string{"json"}

	testExamples(t, cfg, "./data/omitBoth")
}

func TestOmitEmpty(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.Tags = []string{"json"}
	cfg.DisableOmitZero = true

	testExamples(t, cfg, "./data/omitEmpty")
}

func TestOmitNone(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.Tags = []string{"json"}
	cfg.DisableOmitEmpty = true
	cfg.DisableOmitZero = true

	testExamples(t, cfg, "./data/omitNone")
}

func TestOmitZero(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.Tags = []string{"json"}
	cfg.DisableOmitEmpty = true

	testExamples(t, cfg, "./data/omitZero")
}

func TestValidation(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/validation")
}

func TestValidationDisabledReadOnly(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.DisableReadOnlyValidation = true

	testExamples(t, cfg, "./data/validationDisabled/readOnly")
}

func TestDisableCustomTypesForMaps(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.DisableCustomTypesForMaps = true

	testExamples(t, cfg, "./data/disableCustomTypesForMaps")
}

func TestMiscWithDefaults(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/miscWithDefaults")
}

func TestCrossPackage(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.SchemaMappings = []generator.SchemaMapping{
		{
			SchemaID:    "https://example.com/schema",
			PackageName: "github.com/atombender/go-jsonschema/tests/helpers/schema",
			OutputName:  "schema.go",
		},
		{
			SchemaID:    "https://example.com/other",
			PackageName: "github.com/atombender/go-jsonschema/tests/data/crossPackage/other",
			OutputName:  "../other/other.go",
		},
	}
	testExampleFile(t, cfg, "./data/crossPackage/schema/schema.json")
}

func TestCrossPackageAllOf(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.SchemaMappings = []generator.SchemaMapping{
		{
			SchemaID:    "https://example.com/schema",
			PackageName: "github.com/atombender/go-jsonschema/tests/helpers/schema",
			OutputName:  "schema.go",
		},
		{
			SchemaID:    "https://example.com/other",
			PackageName: "github.com/atombender/go-jsonschema/tests/data/crossPackageAllOf/other",
			OutputName:  "../other/other.go",
		},
	}
	testExampleFile(t, cfg, "./data/crossPackageAllOf/schema/schema.json")
}

func TestCrossPackageNoOutput(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.SchemaMappings = []generator.SchemaMapping{
		{
			SchemaID:    "https://example.com/schema",
			PackageName: "github.com/atombender/go-jsonschema/tests/helpers/schema",
			OutputName:  "schema.go",
		},
		{
			SchemaID:    "https://example.com/other",
			PackageName: "github.com/atombender/go-jsonschema/tests/helpers/other",
		},
	}
	testExampleFile(t, cfg, "./data/crossPackageNoOutput/schema/schema.json")
}

func TestBooleanAsSchema(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	testExampleFile(t, cfg, "./data/misc/booleanAsSchema/booleanAsSchema.json")
}

func TestCapitalization(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.Capitalizations = []string{"ID", "URL", "HtMl"}
	testExampleFile(t, cfg, "./data/misc/capitalization/capitalization.json")
}

func TestOnlyModels(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.OnlyModels = true

	testExampleFile(t, cfg, "./data/misc/onlyModels/onlyModels.json")
}

// TestOnlyModelsOneOfPrimitive exercises the OnlyModels fallback for primitive
// `oneOf` schemas: without the gate added in generateDeclaredType, the
// wrapper-emission path would emit a struct whose only field is the
// unexported `value any`, with no methods — unusable to consumers outside
// the generated package. With the gate the schema falls back to the regular
// `interface{}` representation that other consumers can construct directly.
func TestOnlyModelsOneOfPrimitive(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.OnlyModels = true

	testExampleFile(t, cfg, "./data/onlyModels/oneOfPrimitive/oneOfPrimitive.json")
}

func TestSpecialCharacters(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	testExampleFile(t, cfg, "./data/misc/specialCharacters/specialCharacters.json")
}

func TestTags(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.Tags = []string{"yaml"}
	testExampleFile(t, cfg, "./data/misc/tags/tags.json")
}

func TestStructNameFromTitle(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.StructNameFromTitle = true
	testExamples(t, cfg, "./data/nameFromTitle")
}

func TestYamlStructNameFromFile(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	testExampleFile(t, cfg, "./data/yaml/yamlStructNameFromFile/yamlStructNameFromFile.yaml")
}

func TestYamlMultilineDescriptions(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.YAMLExtensions = []string{"yaml"}
	testExampleFile(t, cfg, "./data/yaml/yamlMultilineDescriptions/yamlMultilineDescriptions.yaml")
}

func TestExtraImportsYAML(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.ExtraImports = true
	testExampleFile(t, cfg, "./data/extraImports/gopkgYAMLv3/gopkgYAMLv3.json")
}

func TestRegressions(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/regressions")
}

func TestFormatValidation(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.FormatValidation = generator.FormatValidationConfig{Enabled: true}

	testExamples(t, cfg, "./data/formatValidation")
}

func TestFormatValidationAllowList(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.FormatValidation = generator.FormatValidationConfig{
		Enabled: true,
		// Mixed-case and surrounding whitespace verify the AllowList
		// normalization: shouldValidate trims and lowercases entries so
		// these match the canonical "uuid" / "email" keywords.
		AllowList: []string{"UUID", " email "},
	}

	testExamples(t, cfg, "./data/formatValidationAllowList")
}

func TestStrictAdditionalPropertiesRespectSchema(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.StrictAdditionalProperties = generator.StrictAdditionalPropertiesRespectSchema

	testExamples(t, cfg, "./data/strictAdditionalProperties")
}

func TestStrictAdditionalPropertiesAlways(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.StrictAdditionalProperties = generator.StrictAdditionalPropertiesStrict

	testExamples(t, cfg, "./data/strictAdditionalPropertiesAlways")
}

func TestKnownSchema(t *testing.T) {
	t.Parallel()

	// Pre-load the canonical schema and seed the loader cache keyed by the
	// $id URL the consumer references. The URL is unreachable on principle
	// (no DNS / no listener); a successful generation proves the cache short-
	// circuits before any HTTP fetch is attempted.
	canonical, err := schemas.FromJSONFile("./data/knownSchema/canonical/canonical.json")
	if err != nil {
		t.Fatalf("preload canonical: %v", err)
	}

	cfg := basicConfig
	cfg.Cache = map[string]*schemas.Schema{
		"https://example.com/canonical/v1/canonical.json": canonical,
	}

	testExampleFile(t, cfg, "./data/knownSchema/consumer/consumer.json")
}

// TestKnownSchemaFragmentRef proves the cache short-circuits even when the
// consumer's $ref carries a fragment (#/$defs/...). The cache key is the
// URL without fragment; the loader hands back the pre-loaded *Schema and
// the generator's existing fragment-resolution logic walks into $defs.
func TestKnownSchemaFragmentRef(t *testing.T) {
	t.Parallel()

	canonical, err := schemas.FromJSONFile("./data/knownSchemaFragment/canonical/canonical.json")
	if err != nil {
		t.Fatalf("preload canonical: %v", err)
	}

	cfg := basicConfig
	cfg.Cache = map[string]*schemas.Schema{
		"https://example.com/canonical/v1/canonical-fragment.json": canonical,
	}

	testExampleFile(t, cfg, "./data/knownSchemaFragment/consumer/consumer.json")
}

func TestSchemaPackageWithAlias(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.SchemaMappings = []generator.SchemaMapping{
		{
			SchemaID:    "https://example.com/header",
			PackageName: "github.com/atombender/go-jsonschema/tests/data/schemaPackageAlias/header/v1",
			OutputName:  "../header/v1/header.go",
			ImportAlias: "headerv1",
		},
		{
			SchemaID:    "https://example.com/jobs",
			PackageName: "github.com/atombender/go-jsonschema/tests/data/schemaPackageAlias/jobs/v1",
			OutputName:  "../jobs/v1/jobs.go",
			ImportAlias: "jobsv1",
		},
	}
	testExampleFile(t, cfg, "./data/schemaPackageAlias/consumer/consumer.json")
}

func TestSchemaPackageRejectsInvalidImportAlias(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.SchemaMappings = []generator.SchemaMapping{
		{
			SchemaID:    "https://example.com/schema",
			PackageName: "example.com/foo/v1",
			ImportAlias: "1bad", // starts with digit, not a valid Go identifier
		},
	}

	_, err := generator.New(cfg)
	if err == nil {
		t.Fatal("expected New to reject invalid ImportAlias, got nil")
	}

	if !errors.Is(err, generator.ErrInvalidImportAlias) {
		t.Errorf("expected ErrInvalidImportAlias, got %v", err)
	}
}

// TestSchemaPackageRejectsConflictingImportAlias asserts New() rejects two
// SchemaMappings that bind the same PackageName to different aliases —
// resolveImportAlias would silently pick whichever it iterates over first
// and ignore the other. CR finding from fork PR #15.
func TestSchemaPackageRejectsConflictingImportAlias(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.SchemaMappings = []generator.SchemaMapping{
		{
			SchemaID:    "https://example.com/schemaA",
			PackageName: "example.com/foo/v1",
			ImportAlias: "foov1",
		},
		{
			SchemaID:    "https://example.com/schemaB",
			PackageName: "example.com/foo/v1", // same package
			ImportAlias: "different",          // conflicting alias
		},
	}

	_, err := generator.New(cfg)
	if err == nil {
		t.Fatal("expected New to reject conflicting aliases, got nil")
	}

	if !errors.Is(err, generator.ErrConflictingImportAlias) {
		t.Errorf("expected ErrConflictingImportAlias, got %v", err)
	}
}

// TestSchemaPackageAcceptsRedundantImportAlias confirms the conflict guard
// only fires on DIFFERENT aliases for the same package — two mappings with
// the SAME alias are redundant but legal (both happen to want the same
// override, so resolveImportAlias picks consistently).
func TestSchemaPackageAcceptsRedundantImportAlias(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.SchemaMappings = []generator.SchemaMapping{
		{
			SchemaID:    "https://example.com/schemaA",
			PackageName: "example.com/foo/v1",
			ImportAlias: "foov1",
		},
		{
			SchemaID:    "https://example.com/schemaB",
			PackageName: "example.com/foo/v1",
			ImportAlias: "foov1", // same alias — no conflict
		},
	}

	if _, err := generator.New(cfg); err != nil {
		t.Fatalf("expected redundant-but-matching aliases to be accepted, got %v", err)
	}
}

func TestStrictAdditionalPropertiesRejectsUnknownMode(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.StrictAdditionalProperties = generator.StrictAdditionalPropertiesMode("rstrict") // typo

	_, err := generator.New(cfg)
	if err == nil {
		t.Fatal("expected New to reject unknown StrictAdditionalProperties mode, got nil")
	}

	if !errors.Is(err, generator.ErrInvalidStrictAdditionalPropertiesMode) {
		t.Errorf("expected ErrInvalidStrictAdditionalPropertiesMode, got %v", err)
	}
}

func TestOneOfPrimitive(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/oneOfPrimitive")
}

func TestOneOfDiscriminated(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/oneOfDiscriminated")
}

func TestRecursiveAllOf(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/recursiveAllOf")
}

func TestRootComposition(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/rootComposition")
}

func TestFidelityWarnings(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/fidelityWarnings")
}

// TestConditionalDiscriminator drives generation for the
// allOf+if[const]/then[/else] tagged-union pattern. Walks every fixture
// under tests/data/conditionalDiscriminator and diffs against the sibling
// golden Go file.
func TestConditionalDiscriminator(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/conditionalDiscriminator")
}

func TestExtraImportsYAMLAdditionalProperties(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.ExtraImports = true
	testExampleFile(t, cfg, "./data/extraImports/gopkgYAMLv3AdditionalProperties/gopkgYAMLv3AdditionalProperties.json")
}

func TestMinSizeInt(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.MinSizedInts = true

	testExamples(t, cfg, "./data/minSizedInts")
}

func TestAliasSingleAllOfAnyOfRefs(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.AliasSingleAllOfAnyOfRefs = true

	testExamples(t, cfg, "./data/aliasSingleAllOfAnyOfRefs")
}

func TestSchemaExtensions(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/schemaExtensions")
}

func TestDeeplyNestedMinimalNames(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.MinimalNames = true

	testExamples(t, cfg, "./data/deeplyNested")
}

func TestRefWithOverridesPath(t *testing.T) {
	t.Parallel()

	cfg := basicConfig
	cfg.OnlyModels = true
	cfg.StructNameFromTitle = true

	testExampleFile(t, cfg, "./data/refWithOverridesPath/schema.json")
}

func TestStructWithConstraints(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/structWithConstraints")
}

func testExamples(t *testing.T, cfg generator.Config, dataDir string) {
	t.Helper()

	fileInfos, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, file := range fileInfos {
		if file.IsDir() {
			testExamples(t, cfg, filepath.Join(dataDir, file.Name()))
		}

		if strings.HasSuffix(file.Name(), ".json") {
			fileName := filepath.Join(dataDir, file.Name())
			if strings.HasSuffix(file.Name(), ".FAIL.json") {
				testFailingExampleFile(t, cfg, fileName)
			} else {
				testExampleFile(t, cfg, fileName)
			}
		}
	}
}

func testExampleFile(t *testing.T, cfg generator.Config, fileName string) {
	t.Helper()

	t.Run(titleFromFileName(fileName), func(t *testing.T) {
		t.Parallel()

		g, err := generator.New(cfg)
		if err != nil {
			t.Fatal(err)
		}

		if err := g.DoFile(fileName); err != nil {
			t.Fatal(err)
		}

		sources, err := g.Sources()
		if err != nil {
			t.Fatal(err)
		}

		if len(sources) == 0 {
			t.Fatal("Expected sources to contain something")
		}

		for outputName, source := range sources {
			if outputName == "-" {
				outputName = strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName)) + ".go"
			}

			goldenFileName := filepath.Join(filepath.Dir(fileName), outputName)
			t.Logf("Using golden data in %s", mustAbs(goldenFileName))

			goldenData, err := os.ReadFile(goldenFileName)
			if err != nil {
				if !os.IsNotExist(err) {
					t.Fatal(err)
				}

				goldenData = source

				t.Log("File does not exist; creating it")

				if err = os.WriteFile(goldenFileName, goldenData, 0o655); err != nil {
					t.Fatal(err)
				}
			}

			// Overwriting the expected file is useful if there are lots of differences
			// due to a code change you made and you just want to accept the new output.
			// Simply run "OVERWRITE_EXPECTED_GO_FILE=true make test".
			if os.Getenv("OVERWRITE_EXPECTED_GO_FILE") == "true" {
				t.Logf("Updating file %s", mustAbs(goldenFileName))

				if err = os.WriteFile(goldenFileName, source, 0o655); err != nil {
					t.Fatalf("Failed to write to %s: %s\n", goldenFileName, err.Error())
				}
			} else {
				if diff := cmp.Diff(string(goldenData), string(source)); diff != "" {
					t.Errorf("Contents different (left is expected, right is actual):\n%s", diff)
				}

				if diff, ok := diffStrings(t, string(goldenData), string(source)); !ok {
					t.Fatalf("Contents different (left is expected, right is actual):\n%s", *diff)
				}
			}
		}
	})
}

func testFailingExampleFile(t *testing.T, cfg generator.Config, fileName string) {
	t.Helper()

	t.Run(titleFromFileName(fileName), func(t *testing.T) {
		g, err := generator.New(cfg)
		if err != nil {
			t.Fatal(err)
		}

		if err := g.DoFile(fileName); err == nil {
			t.Fatal("Expected test to fail")
		}
	})
}

func diffStrings(t *testing.T, expected, actual string) (*string, bool) {
	t.Helper()

	if actual == expected {
		return nil, true
	}

	dir := t.TempDir()

	if err := os.WriteFile(fmt.Sprintf("%s/expected", dir), []byte(expected), 0o644); err != nil {
		t.Fatal(err.Error())
	}

	if err := os.WriteFile(fmt.Sprintf("%s/actual", dir), []byte(actual), 0o644); err != nil {
		t.Fatal(err.Error())
	}

	out, err := exec.CommandContext(
		context.Background(),
		"diff",
		"--side-by-side",
		fmt.Sprintf("%s/expected", dir),
		fmt.Sprintf("%s/actual", dir),
	).Output()

	if !errors.As(err, &errExit) {
		t.Fatal(err.Error())
	}

	diff := string(out)

	return &diff, false
}

func titleFromFileName(fileName string) string {
	relative := mustRel(mustAbs("./data"), mustAbs(fileName))

	return strings.TrimSuffix(relative, ".json")
}

func mustRel(base, s string) string {
	result, err := filepath.Rel(base, s)
	if err != nil {
		panic(err)
	}

	return result
}

func mustAbs(s string) string {
	result, err := filepath.Abs(s)
	if err != nil {
		panic(err)
	}

	return result
}
