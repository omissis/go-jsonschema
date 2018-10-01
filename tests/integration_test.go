package tests

import (
	"fmt"
	"io/ioutil"
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
	Warner:             func(message string) {},
}

func TestCore(t *testing.T) {
	generator, err := generator.New(basicConfig)
	if err != nil {
		t.Error(err)
	}
	testExamples(t, generator, "./data/core")
}

func TestValidation(t *testing.T) {
	generator, err := generator.New(basicConfig)
	if err != nil {
		t.Error(err)
	}
	testExamples(t, generator, "./data/validation")
}

func TestMiscWithDefaults(t *testing.T) {
	generator, err := generator.New(basicConfig)
	if err != nil {
		t.Error(err)
	}
	testExamples(t, generator, "./data/miscWithDefaults")
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
	generator, err := generator.New(cfg)
	if err != nil {
		t.Error(err)
	}
	testExampleFile(t, generator, "./data/crossPackage/schema.json")
}

func TestCapitalization(t *testing.T) {
	cfg := basicConfig
	cfg.Capitalizations = []string{"ID", "URL", "HtMl"}
	generator, err := generator.New(cfg)
	if err != nil {
		t.Error(err)
	}
	testExampleFile(t, generator, "./data/misc/capitalization.json")
}

func testExamples(t *testing.T, generator *generator.Generator, dataDir string) {
	fileInfos, err := ioutil.ReadDir(dataDir)
	if err != nil {
		t.Error(err.Error())
	}

	for _, file := range fileInfos {
		if strings.HasSuffix(file.Name(), ".json") {
			fileName := filepath.Join(dataDir, file.Name())
			if strings.HasSuffix(file.Name(), ".FAIL.json") {
				testFailingExampleFile(t, generator, fileName)
			} else {
				testExampleFile(t, generator, fileName)
			}
		}
	}
}

func testExampleFile(t *testing.T, generator *generator.Generator, fileName string) {
	t.Run(fileName, func(t *testing.T) {
		if err := generator.DoFile(fileName); err != nil {
			t.Error(err)
		}

		if len(generator.Sources()) == 0 {
			t.Error("Expected sources to contain something")
		}

		for outputName, source := range generator.Sources() {
			if outputName == "-" {
				outputName = strings.TrimSuffix(filepath.Base(fileName), ".json") + ".go"
			}
			outputName += ".output"

			goldenFileName := filepath.Join(filepath.Dir(fileName), outputName)

			goldenData, err := ioutil.ReadFile(goldenFileName)
			if err != nil {
				if !os.IsNotExist(err) {
					t.Error(err)
				}
				goldenData = source
				t.Logf("Writing golden data to %s", goldenFileName)
				if err = ioutil.WriteFile(goldenFileName, goldenData, 0655); err != nil {
					t.Error(err)
				}
			}
			if diff, ok := diffStrings(t, string(goldenData), string(source)); !ok {
				t.Error(fmt.Sprintf("Contents different (left is expected, right is actual):\n%s", *diff))
			}
		}
	})
}

func testFailingExampleFile(t *testing.T, generator *generator.Generator, fileName string) {
	t.Run(fileName, func(t *testing.T) {
		if err := generator.DoFile(fileName); err == nil {
			t.Error("Expected test to fail")
		}
	})
}

func diffStrings(t *testing.T, expected, actual string) (*string, bool) {
	if actual == expected {
		return nil, true
	}

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err.Error())
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	if err := ioutil.WriteFile(fmt.Sprintf("%s/expected", dir), []byte(expected), 0644); err != nil {
		t.Error(err.Error())
	}
	if err := ioutil.WriteFile(fmt.Sprintf("%s/actual", dir), []byte(actual), 0644); err != nil {
		t.Error(err.Error())
	}

	out, err := exec.Command("diff", "--side-by-side",
		fmt.Sprintf("%s/expected", dir),
		fmt.Sprintf("%s/actual", dir)).Output()
	if _, ok := err.(*exec.ExitError); !ok {
		t.Error(err.Error())
	}

	diff := string(out)
	return &diff, false
}
