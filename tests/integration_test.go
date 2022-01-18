package tests

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/atombender/go-jsonschema/pkg/generator"
)

var basicConfig = generator.Config{
	SchemaMappings:     []generator.SchemaMapping{},
	DefaultPackageName: "github.com/example/test",
	DefaultOutputName:  "-",
	ResolveExtensions:  []string{".json", ".yaml"},
	Warner: func(message string) {
		log.Printf("[from warner] %s", message)
	},
}

func TestCore(t *testing.T) {
	testExamples(t, basicConfig, "./data/core")
}

func TestValidation(t *testing.T) {
	testExamples(t, basicConfig, "./data/validation")
}

func TestMiscWithDefaults(t *testing.T) {
	testExamples(t, basicConfig, "./data/miscWithDefaults")
}

func TestCrossPackage(t *testing.T) {
	cfg := basicConfig
	cfg.SchemaMappings = []generator.SchemaMapping{
		{
			SchemaID:    "https://example.com/schema",
			PackageName: "github.com/example/schema",
			OutputName:  "schema.go",
		},
		{
			SchemaID:    "https://example.com/other",
			PackageName: "github.com/example/other",
			OutputName:  "other.go",
		},
	}
	testExampleFile(t, cfg, "./data/crossPackage/schema.json")
}

func TestCrossPackageNoOutput(t *testing.T) {
	cfg := basicConfig
	cfg.SchemaMappings = []generator.SchemaMapping{
		{
			SchemaID:    "https://example.com/schema",
			PackageName: "github.com/example/schema",
			OutputName:  "schema.go",
		},
		{
			SchemaID:    "https://example.com/other",
			PackageName: "github.com/example/other",
		},
	}
	testExampleFile(t, cfg, "./data/crossPackageNoOutput/schema.json")
}

func TestCapitalization(t *testing.T) {
	cfg := basicConfig
	cfg.Capitalizations = []string{"ID", "URL", "HtMl"}
	testExampleFile(t, cfg, "./data/misc/capitalization.json")
}

func TestBooleanAsSchema(t *testing.T) {
	cfg := basicConfig
	testExampleFile(t, cfg, "./data/misc/boolean-as-schema.json")
}

func TestYamlMultilineDescriptions(t *testing.T) {
	cfg := basicConfig
	cfg.YAMLExtensions = []string{"yaml"}
	testExampleFile(t, cfg, "./data/yaml/yamlMultilineDescriptions.yaml")
}

func testExamples(t *testing.T, cfg generator.Config, dataDir string) {
	fileInfos, err := ioutil.ReadDir(dataDir)
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, file := range fileInfos {
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
	t.Run(titleFromFileName(fileName), func(t *testing.T) {
		generator, err := generator.New(cfg)
		if err != nil {
			t.Fatal(err)
		}

		if err := generator.DoFile(fileName); err != nil {
			t.Fatal(err)
		}

		if len(generator.Sources()) == 0 {
			t.Fatal("Expected sources to contain something")
		}

		for outputName, source := range generator.Sources() {
			if outputName == "-" {
				outputName = strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName)) + ".go"
			}
			outputName += ".output"

			goldenFileName := filepath.Join(filepath.Dir(fileName), outputName)
			t.Logf("Using golden data in %s", mustAbs(goldenFileName))

			goldenData, err := ioutil.ReadFile(goldenFileName)
			if err != nil {
				if !os.IsNotExist(err) {
					t.Fatal(err)
				}
				goldenData = source
				t.Log("File does not exist; creating it")
				if err = ioutil.WriteFile(goldenFileName, goldenData, 0655); err != nil {
					t.Fatal(err)
				}
			}
			if diff, ok := diffStrings(t, string(goldenData), string(source)); !ok {
				t.Fatal(fmt.Sprintf("Contents different (left is expected, right is actual):\n%s", *diff))
			}
		}
	})
}

func testFailingExampleFile(t *testing.T, cfg generator.Config, fileName string) {
	t.Run(titleFromFileName(fileName), func(t *testing.T) {
		generator, err := generator.New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if err := generator.DoFile(fileName); err == nil {
			t.Fatal("Expected test to fail")
		}
	})
}

func diffStrings(t *testing.T, expected, actual string) (*string, bool) {
	if actual == expected {
		return nil, true
	}

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	if err := ioutil.WriteFile(fmt.Sprintf("%s/expected", dir), []byte(expected), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := ioutil.WriteFile(fmt.Sprintf("%s/actual", dir), []byte(actual), 0644); err != nil {
		t.Fatal(err.Error())
	}

	out, err := exec.Command("diff", "--side-by-side",
		fmt.Sprintf("%s/expected", dir),
		fmt.Sprintf("%s/actual", dir)).Output()
	if _, ok := err.(*exec.ExitError); !ok {
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
