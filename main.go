package main

import (
	"errors"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atombender/go-jsonschema/pkg/generator"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

const (
	perm644 = 0o644
	perm755 = 0o755

	// flagValueOff is the explicit-off literal accepted by the opt-in flags
	// (--strict-additional-properties, --validate-formats). Centralized so the
	// linter (goconst) sees a single canonical reference.
	flagValueOff = "off"
)

var (
	verbose                       bool
	extraImports                  bool
	onlyModels                    bool
	defaultPackage                string
	defaultOutput                 string
	schemaPackages                []string
	schemaOutputs                 []string
	schemaRootTypes               []string
	knownSchemas                  []string
	capitalizations               []string
	resolveExtensions             []string
	yamlExtensions                []string
	tags                          []string
	structNameFromTitle           bool
	minSizedInts                  bool
	minimalNames                  bool
	disableReadOnlyValidation     bool
	disableCustomTypesForMaps     bool
	disableOmitEmpty              bool
	disableOmitZero               bool
	validateFormatsRaw            string
	strictAdditionalPropertiesRaw string

	errFlagFormat            = errors.New("flag must be in the format URI=PACKAGE")
	errUnknownFormatKeyword  = errors.New("unknown format keyword")
	errEmptyFormatListEntry  = errors.New("empty format name in --validate-formats list")
	errFormatListWithKeyword = errors.New(
		`--validate-formats: "all" and "off" cannot be combined with format names`,
	)
	errInvalidStrictAddlPropMode = errors.New(
		"--strict-additional-properties: invalid mode (expected one of: off, respect-schema, strict)",
	)
	errInvalidImportAlias = errors.New(
		"--schema-package: text after the final ':' must be a valid Go identifier" +
			" (use the library API if your package path legitimately contains ':')",
	)
	errKnownSchemaLoad      = errors.New("--known-schema: failed to load file")
	errKnownSchemaDuplicate = errors.New("--known-schema: duplicate URL")

	rootCmd = &cobra.Command{
		Use:   "go-jsonschema FILE ...",
		Short: "Generates Go code from JSON Schema files.",
		Run: func(_ *cobra.Command, args []string) {
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

			formatValidation, err := parseValidateFormats(validateFormatsRaw)
			if err != nil {
				abortWithErr(err)
			}

			strictAddlProps, err := parseStrictAdditionalProperties(strictAdditionalPropertiesRaw)
			if err != nil {
				abortWithErr(err)
			}

			knownSchemaCache, err := loadKnownSchemas(knownSchemas, yamlExtensions)
			if err != nil {
				abortWithErr(err)
			}

			cfg := generator.Config{
				Warner: func(message string) {
					logf("Warning: %s", message)
				},
				ExtraImports:               extraImports,
				Capitalizations:            capitalizations,
				DefaultOutputName:          defaultOutput,
				DefaultPackageName:         defaultPackage,
				SchemaMappings:             []generator.SchemaMapping{},
				ResolveExtensions:          resolveExtensions,
				YAMLExtensions:             yamlExtensions,
				StructNameFromTitle:        structNameFromTitle,
				Tags:                       tags,
				OnlyModels:                 onlyModels,
				MinSizedInts:               minSizedInts,
				MinimalNames:               minimalNames,
				DisableReadOnlyValidation:  disableReadOnlyValidation,
				DisableCustomTypesForMaps:  disableCustomTypesForMaps,
				DisableOmitEmpty:           disableOmitEmpty,
				DisableOmitZero:            disableOmitZero,
				FormatValidation:           formatValidation,
				StrictAdditionalProperties: strictAddlProps,
				Cache:                      knownSchemaCache,
			}

			for _, id := range allKeys(schemaPackageMap, schemaOutputMap, schemaRootTypeMap) {
				mapping := generator.SchemaMapping{SchemaID: id}
				if s, ok := schemaPackageMap[id]; ok {
					pkg, alias, err := splitPackageAlias(s)
					if err != nil {
						abortWithErr(err)
					}

					mapping.PackageName = pkg
					mapping.ImportAlias = alias
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

				cfg.SchemaMappings = append(cfg.SchemaMappings, mapping)
			}

			generator, err := generator.New(cfg)
			if err != nil {
				abortWithErr(err)
			}

			for _, fileName := range args {
				verboseLogf("Loading %s", fileName)

				if err = generator.DoFile(fileName); err != nil {
					abortWithErr(err)
				}
			}

			sources, err := generator.Sources()
			if err != nil {
				abortWithErr(err)
			}

			for fileName, source := range sources {
				if fileName != "-" {
					verboseLogf("Writing %s", fileName)
				}

				if fileName == "-" {
					if _, err = os.Stdout.Write(source); err != nil {
						abortWithErr(err)
					}
				} else {
					if err := os.MkdirAll(filepath.Dir(fileName), perm755); err != nil {
						abortWithErr(err)
					}

					w, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm644)
					if err != nil {
						abortWithErr(err)
					}

					if _, err := w.Write(source); err != nil {
						abortWithErr(err)
					}

					if err := w.Close(); err != nil {
						abortWithErr(err)
					}
				}
			}

			os.Exit(0)
		},
	}
)

func main() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Verbose output")
	rootCmd.PersistentFlags().BoolVarP(&extraImports, "extra-imports", "e", false,
		"Allow extra imports (non standard library)")
	rootCmd.PersistentFlags().BoolVar(&onlyModels, "only-models", false,
		"Generate only models (no unmarshal methods, no validation)")
	rootCmd.PersistentFlags().StringVarP(&defaultPackage, "package", "p", "",
		`Default name of package to declare Go files under, unless overridden with
--schema-package`)
	rootCmd.PersistentFlags().StringVarP(&defaultOutput, "output", "o", "-",
		"File to write (- for standard output)")
	rootCmd.PersistentFlags().StringSliceVar(&schemaPackages, "schema-package", nil,
		`Name of package to declare Go files for a specific schema ID under;
must be in the format URI=PACKAGE or URI=PACKAGE:ALIAS. The optional :ALIAS
suffix overrides the import alias used in generated code (defaults to the
package path's last segment); set it when two referenced packages share a
last segment (e.g. both end in /v1) to avoid an import-alias collision.`)
	rootCmd.PersistentFlags().StringSliceVar(&knownSchemas, "known-schema", nil,
		`Pre-populate the loader cache from a local file so that any $ref to URL
resolves to the file's contents without an HTTP fetch (or, for URLs with
file:// or relative scheme, without a disk lookup beyond this entry); must
be in the format URL=PATH. The schema is loaded for ref resolution only —
no Go code is emitted for it. Repeat for each cross-tree dependency.`)
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
	rootCmd.PersistentFlags().StringSliceVar(&yamlExtensions, "yaml-extension", []string{".yml", ".yaml"},
		`Add a file extension that should be recognized as YAML. Default are .yml, .yaml.`)
	rootCmd.PersistentFlags().BoolVarP(&structNameFromTitle, "struct-name-from-title", "t", false,
		"Use the schema title as the generated struct name")
	rootCmd.PersistentFlags().StringSliceVar(&tags, "tags", []string{"json", "yaml", "mapstructure"},
		`Specify which struct tags to generate. Defaults are json, yaml, mapstructure`)
	rootCmd.PersistentFlags().BoolVar(&minSizedInts, "min-sized-ints", false,
		"Uses sized int and uint values based on the min and max values for the field")
	rootCmd.PersistentFlags().BoolVar(&disableReadOnlyValidation, "disable-readonly-validation", false,
		"Do not include validation of readonly fields")
	rootCmd.PersistentFlags().BoolVar(&minimalNames, "minimal-names", false,
		"Uses the shortest possible names")
	rootCmd.PersistentFlags().BoolVar(&disableCustomTypesForMaps, "disable-custom-types-for-maps", false,
		"Do not generate custom types when generating maps")
	rootCmd.PersistentFlags().BoolVar(&disableOmitEmpty, "disable-omitempty", false,
		"disable the addition of omitempty tag values")
	rootCmd.PersistentFlags().BoolVar(&disableOmitZero, "disable-omitzero", false,
		"disable the addition of omitzero tag values")
	rootCmd.PersistentFlags().StringVar(&validateFormatsRaw, "validate-formats", "",
		`Opt in to runtime validation of JSON Schema "format" keywords. Accepts
"off" (default; same as omitting the flag), "all" (validate every supported
format), or a comma-separated subset (e.g. "uuid,email"). Supported formats:
`+strings.Join(generator.SupportedFormats(), ", ")+`.`)
	rootCmd.PersistentFlags().StringVar(&strictAdditionalPropertiesRaw, "strict-additional-properties", "",
		`Opt in to runtime rejection of unknown fields. Accepts "off" (default;
same as omitting the flag), "respect-schema" (reject unknown fields only for
objects whose schema declares additionalProperties: false), or "strict"
(reject unknown fields for every generated object type — except when the
schema declares a typed additionalProperties, which generates a catch-all
map field instead, or when patternProperties is present, which suppresses
enforcement with a warning).`)

	abortWithErr(rootCmd.Execute())
}

// parseStrictAdditionalProperties interprets the --strict-additional-properties
// flag value. Accepted values map directly to the generator package's three
// modes; an empty string means the flag was not set and behaves like "off".
func parseStrictAdditionalProperties(raw string) (generator.StrictAdditionalPropertiesMode, error) {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" || trimmed == flagValueOff {
		return generator.StrictAdditionalPropertiesOff, nil
	}

	mode := generator.StrictAdditionalPropertiesMode(trimmed)
	if !mode.IsValid() {
		return "", fmt.Errorf("%w: %q", errInvalidStrictAddlPropMode, raw)
	}

	return mode, nil
}

// parseValidateFormats interprets the --validate-formats flag value.
//
// Accepted values:
//
//	""         — flag omitted; validation off.
//	"off"      — explicit off (useful for scripts that always set the flag).
//	"all"      — validate every format returned by generator.SupportedFormats.
//	"a,b,c"    — validate only the listed formats (case-insensitive).
//
// Any unknown format name is rejected up front rather than silently dropped,
// so a typo like "uuid,emial" fails immediately instead of disabling email
// validation. "all" and "off" cannot be mixed with format names.
func parseValidateFormats(raw string) (generator.FormatValidationConfig, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.EqualFold(trimmed, flagValueOff) {
		return generator.FormatValidationConfig{}, nil
	}

	if strings.EqualFold(trimmed, "all") {
		return generator.FormatValidationConfig{Enabled: true}, nil
	}

	parts := strings.Split(trimmed, ",")
	allowList := make([]string, 0, len(parts))

	for _, p := range parts {
		name := strings.ToLower(strings.TrimSpace(p))
		switch {
		case name == "":
			return generator.FormatValidationConfig{}, errEmptyFormatListEntry
		case name == "all" || name == flagValueOff:
			return generator.FormatValidationConfig{}, errFormatListWithKeyword
		case !generator.IsSupportedFormat(name):
			return generator.FormatValidationConfig{}, fmt.Errorf(
				"%w %q; supported: %s",
				errUnknownFormatKeyword, name, strings.Join(generator.SupportedFormats(), ", "),
			)
		}

		allowList = append(allowList, name)
	}

	return generator.FormatValidationConfig{Enabled: true, AllowList: allowList}, nil
}

func abortWithErr(err error) {
	if err != nil {
		abort(err.Error())
	}
}

func abort(message string) {
	logf("Failed: %s", message)
	os.Exit(1)
}

func stringSliceToStringMap(s []string) (map[string]string, error) {
	result := make(map[string]string, len(s))

	for _, p := range s {
		i := strings.IndexRune(p, '=')
		if i == -1 {
			return nil, fmt.Errorf("%w: %q", errFlagFormat, p)
		}

		result[p[0:i]] = p[i+1:]
	}

	return result, nil
}

// loadKnownSchemas parses each --known-schema URL=PATH entry into a
// pre-populated loader cache. Files with extensions in the YAML set are
// parsed with FromYAMLFile; everything else with FromJSONFile, mirroring
// FileLoader.parseFile. Duplicate URLs are rejected at flag-parse time
// rather than silently overwriting (last-write-wins would mask copy/paste
// mistakes in long flag lists).
func loadKnownSchemas(entries, yamlExtensions []string) (map[string]*schemas.Schema, error) {
	cache := make(map[string]*schemas.Schema, len(entries))
	if len(entries) == 0 {
		return cache, nil
	}

	yamlExtSet := map[string]bool{}
	for _, ext := range yamlExtensions {
		yamlExtSet[ext] = true
	}

	for _, entry := range entries {
		i := strings.IndexRune(entry, '=')
		if i == -1 {
			return nil, fmt.Errorf("%w: %q", errFlagFormat, entry)
		}

		url, path := entry[0:i], entry[i+1:]
		if _, exists := cache[url]; exists {
			return nil, fmt.Errorf("%w: %q", errKnownSchemaDuplicate, url)
		}

		var (
			sc  *schemas.Schema
			err error
		)
		if yamlExtSet[filepath.Ext(path)] {
			sc, err = schemas.FromYAMLFile(path)
		} else {
			sc, err = schemas.FromJSONFile(path)
		}

		if err != nil {
			return nil, fmt.Errorf("%w %q: %w", errKnownSchemaLoad, path, err)
		}

		cache[url] = sc
	}

	return cache, nil
}

// splitPackageAlias parses the value side of a --schema-package mapping.
// Grammar: PACKAGE[:ALIAS] (split on the LAST colon). When the suffix is
// present, it must be a valid non-keyword Go identifier; otherwise the
// caller gets errInvalidImportAlias and is told to use the library API
// for the rare case of a package path that legitimately contains ':'.
// When no colon is present, alias is empty and the entire input is the
// package path (the historical behavior).
func splitPackageAlias(value string) (string, string, error) {
	i := strings.LastIndex(value, ":")
	if i == -1 {
		return value, "", nil
	}

	candidate := value[i+1:]
	if candidate == "" || token.IsKeyword(candidate) || !token.IsIdentifier(candidate) {
		return "", "", fmt.Errorf("%w: %q", errInvalidImportAlias, candidate)
	}

	return value[:i], candidate, nil
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

func logf(format string, args ...any) {
	fmt.Fprint(os.Stderr, "go-jsonschema: ")
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprint(os.Stderr, "\n")
}

func verboseLogf(format string, args ...any) {
	if verbose {
		logf(format, args...)
	}
}
