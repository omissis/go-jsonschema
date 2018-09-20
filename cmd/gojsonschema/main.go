package main

import (
	"fmt"
	"go/format"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atombender/go-jsonschema/pkg/generator"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

var (
	verbose         bool
	defaultPackage  string
	defaultOutput   string
	schemaPackages  []string
	schemaOutputs   []string
	schemaRootTypes []string
)

var rootCmd = &cobra.Command{
	Use:   "gojsonschema FILE ...",
	Short: "Generates Go code from JSON Schema files.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			abort("No arguments specified. Run with --help for usage.")
		}

		if defaultPackage == "" && len(schemaPackages) == 0 {
			abort("Package name not specified.")
		}

		schemaPackageMap, err := stringSliceToStringMap(schemaPackages)
		if err != nil {
			abortWithErr(err)
		}

		schemaOutputMap, err := stringSliceToStringMap(schemaOutputs)
		if err != nil {
			abortWithErr(err)
		}

		schemaRootTypeMap, err := stringSliceToStringMap(schemaRootTypes)
		if err != nil {
			abortWithErr(err)
		}

		schemaMappings := []generator.SchemaMapping{}
		for _, id := range allKeys(schemaPackageMap, schemaOutputMap, schemaRootTypeMap) {
			mapping := generator.SchemaMapping{SchemaID: id}
			if s, ok := schemaPackageMap[id]; ok {
				mapping.PackageName = s
			} else {
				mapping.PackageName = defaultPackage
			}
			if s, ok := schemaOutputMap[id]; ok {
				mapping.OutputName = s
			} else {
				mapping.OutputName = defaultOutput
			}
			if s, ok := schemaRootTypeMap[id]; ok {
				mapping.RootType = s
			}
			schemaMappings = append(schemaMappings, mapping)
		}

		generator, err := generator.New(schemaMappings, defaultPackage, defaultOutput, func(message string) {
			log("Warning: %s", message)
		})
		if err != nil {
			abortWithErr(err)
		}

		for _, fileName := range args {
			verboseLog("Loading %s", fileName)

			f, err := os.Open(fileName)
			if err != nil {
				abortWithErr(err)
			}
			defer func() { _ = f.Close() }()

			schema, err := schemas.FromReader(f)
			if err != nil {
				abortWithErr(err)
			}
			_ = f.Close()

			if err = generator.AddFile(fileName, schema); err != nil {
				abortWithErr(err)
			}
		}

		for fileName, source := range generator.Sources() {
			if fileName != "-" {
				verboseLog("Writing %s", fileName)
			}

			src, err := format.Source(source)
			if err != nil {
				verboseLog("the generated code could not be formatted automatically; "+
					"falling back to unformatted: %s", err)
				src = source
			}

			if fileName == "-" {
				if _, err = os.Stdout.Write(src); err != nil {
					abortWithErr(err)
				}
			} else {
				w, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					abortWithErr(err)
				}
				defer func() { _ = w.Close() }()
				if _, err = w.Write(src); err != nil {
					abortWithErr(err)
				}
				_ = w.Close()
			}
		}

		os.Exit(0)
	},
}

func main() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Verbose output")
	rootCmd.PersistentFlags().StringVarP(&defaultPackage, "package", "p", "",
		"Default name of package to declare Go files under, unless overridden with --schema-package")
	rootCmd.PersistentFlags().StringVarP(&defaultOutput, "output", "o", "-",
		"File to write (- for standard output)")
	rootCmd.PersistentFlags().StringSliceVar(&schemaPackages, "schema-package", nil,
		"Name of package to declare Go files for a specific schema ID under; "+
			"must be in the format URI=PACKAGE.")
	rootCmd.PersistentFlags().StringSliceVar(&schemaOutputs, "schema-output", nil,
		"File to write (- for standard output) a specific schema ID to; "+
			"must be in the format URI=PACKAGE.")
	rootCmd.PersistentFlags().StringSliceVar(&schemaRootTypes, "schema-root-type", nil,
		"Override name to use for the root type of a specific schema ID; "+
			"must be in the format URI=PACKAGE. By default, it is derived from the file name.")

	abortWithErr(rootCmd.Execute())
}

func abortWithErr(err error) {
	if err != nil {
		abort(err.Error())
	}
}

func abort(message string) {
	log("Failed: %s", message)
	os.Exit(1)
}

func stringSliceToStringMap(s []string) (map[string]string, error) {
	result := make(map[string]string, len(s))
	for _, p := range s {
		i := strings.IndexRune(p, '=')
		if i == -1 {
			return nil, fmt.Errorf("flag must be in the format URI=PACKAGE: %q", p)
		}
		result[p[0:i]] = p[i+1:]
	}
	return result, nil
}

func allKeys(in ...map[string]string) []string {
	type dummy struct{}
	keySet := map[string]dummy{}
	for _, m := range in {
		for k := range m {
			keySet[k] = dummy{}
		}
	}
	result := make([]string, 0, len(keySet))
	for k := range keySet {
		result = append(result, k)
	}
	return result
}

func log(format string, args ...interface{}) {
	fmt.Fprint(os.Stderr, "gojsonschema: ")
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprint(os.Stderr, "\n")
}

func verboseLog(format string, args ...interface{}) {
	if verbose {
		log(format, args...)
	}
}
