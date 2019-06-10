package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atombender/go-jsonschema/pkg/generator"
)

var (
	verbose           bool
	defaultPackage    string
	defaultOutput     string
	schemaPackages    []string
	schemaOutputs     []string
	schemaRootTypes   []string
	capitalizations   []string
	resolveExtensions []string
	yamlExtensions    = []string{".yml", ".yaml"}
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

		cfg := generator.Config{
			Warner: func(message string) {
				log("Warning: %s", message)
			},
			Capitalizations:    capitalizations,
			DefaultOutputName:  defaultOutput,
			DefaultPackageName: defaultPackage,
			SchemaMappings:     []generator.SchemaMapping{},
			ResolveExtensions:  resolveExtensions,
			YAMLExtensions:     yamlExtensions,
		}
		for _, id := range allKeys(schemaPackageMap, schemaOutputMap, schemaRootTypeMap) {
			mapping := generator.SchemaMapping{SchemaID: id}
			if s, ok := schemaPackageMap[id]; ok {
				mapping.PackageName = s
			} else {
				mapping.PackageName = defaultPackage
			}
			if s, ok := schemaOutputMap[id]; ok {
				mapping.OutputName = s
			}
			if s, ok := schemaRootTypeMap[id]; ok {
				mapping.RootType = s
			}
			cfg.SchemaMappings = append(cfg.SchemaMappings, mapping)
		}

		generator, err := generator.New(cfg)
		if err != nil {
			abortWithErr(err)
		}

		for _, fileName := range args {
			verboseLog("Loading %s", fileName)
			if err = generator.DoFile(fileName); err != nil {
				abortWithErr(err)
			}
		}

		for fileName, source := range generator.Sources() {
			if fileName != "-" {
				verboseLog("Writing %s", fileName)
			}

			if fileName == "-" {
				if _, err = os.Stdout.Write(source); err != nil {
					abortWithErr(err)
				}
			} else {
				if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
					abortWithErr(err)
				}
				w, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					abortWithErr(err)
				}
				if _, err = w.Write(source); err != nil {
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
		`Default name of package to declare Go files under, unless overridden with
--schema-package`)
	rootCmd.PersistentFlags().StringVarP(&defaultOutput, "output", "o", "-",
		"File to write (- for standard output)")
	rootCmd.PersistentFlags().StringSliceVar(&schemaPackages, "schema-package", nil,
		`Name of package to declare Go files for a specific schema ID under;
must be in the format URI=PACKAGE.`)
	rootCmd.PersistentFlags().StringSliceVar(&schemaOutputs, "schema-output", nil,
		`File to write (- for standard output) a specific schema ID to;
must be in the format URI=FILENAME.`)
	rootCmd.PersistentFlags().StringSliceVar(&schemaRootTypes, "schema-root-type", nil,
		`Override name to use for the root type of a specific schema ID;
must be in the format URI=TYPE. By default, it is derived from the file name.`)
	rootCmd.PersistentFlags().StringSliceVar(&capitalizations, "capitalization", nil,
		`Specify a preferred Go capitalization for a string. For example, by default a field
named 'id' becomes 'Id'. With --capitalization ID, it will be generated as 'ID'.`)
	rootCmd.PersistentFlags().StringSliceVar(&resolveExtensions, "resolve-extension", nil,
		`Add a file extension that is used to resolve schema names, e.g. {"$ref": "./foo"} will
also look for foo.json if --resolve-extension json is provided.`)
	rootCmd.PersistentFlags().StringSliceVar(&yamlExtensions, "yaml-extension", nil,
		`Add a file extension that should be recognized as YAML. Default are .yml, .yaml.`)

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
