package tests_test

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/atombender/go-jsonschema/pkg/generator"
)

var (
	exitErr *exec.ExitError

	basicConfig = generator.Config{
		SchemaMappings:     []generator.SchemaMapping{},
		ExtraImports:       false,
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

func TestCore(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/core")
}

func TestValidation(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/validation")
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

func TestRegressions(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/regressions")
}

func TestSchemaExtensions(t *testing.T) {
	t.Parallel()

	testExamples(t, basicConfig, "./data/schemaExtensions")
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

		if len(g.Sources()) == 0 {
			t.Fatal("Expected sources to contain something")
		}

		for outputName, source := range g.Sources() {
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

			if diff, ok := diffStrings(t, string(goldenData), string(source)); !ok {
				t.Fatalf("Contents different (left is expected, right is actual):\n%s", *diff)
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

	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err.Error())
	}

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	if err = os.WriteFile(fmt.Sprintf("%s/expected", dir), []byte(expected), 0o644); err != nil {
		t.Fatal(err.Error())
	}

	if err = os.WriteFile(fmt.Sprintf("%s/actual", dir), []byte(actual), 0o644); err != nil {
		t.Fatal(err.Error())
	}

	out, err := exec.Command("diff", "--side-by-side",
		fmt.Sprintf("%s/expected", dir),
		fmt.Sprintf("%s/actual", dir)).Output()

	if !errors.As(err, &exitErr) {
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
