package tests

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/atombender/go-jsonschema/pkg/generator"
)

func TestCore(t *testing.T) {
	generator, err := generator.New([]generator.SchemaMapping{},
		"github.com/example/test", "-", func(message string) {})
	if err != nil {
		t.Error(err)
	}
	testExamples(t, generator, "./data/core")
}

func TestValidation(t *testing.T) {
	generator, err := generator.New([]generator.SchemaMapping{},
		"github.com/example/test", "-", func(message string) {})
	if err != nil {
		t.Error(err)
	}
	testExamples(t, generator, "./data/validation")
}

func TestMisc(t *testing.T) {
	generator, err := generator.New([]generator.SchemaMapping{},
		"github.com/example/test", "-", func(message string) {})
	if err != nil {
		t.Error(err)
	}
	testExamples(t, generator, "./data/misc")
}

func TestCrossPackage(t *testing.T) {
	generator, err := generator.New([]generator.SchemaMapping{
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
	}, "github.com/example/test", "-", func(message string) {})
	if err != nil {
		t.Error(err)
	}
	testExampleFile(t, generator, "./data/crossPackage/schema.json")
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
			if !reflect.DeepEqual(source, goldenData) {
				t.Logf("Expected: %s", goldenData)
				t.Logf("Actual: %s", source)
				t.Error("Contents different")
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
