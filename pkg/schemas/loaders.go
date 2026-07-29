package schemas

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	ErrCannotResolveSchema      = errors.New("cannot resolve schema")
	ErrCannotLoadSchema         = errors.New("cannot load schema")
	ErrUnsupportedContentType   = errors.New("unsupported content type")
	ErrUnsupportedFileExtension = errors.New("unsupported file extension")
	ErrUnsupportedURL           = errors.New("unsupported URL")
)

type Loader interface {
	Load(uri, parentURI string) (*Schema, error)
}

// RefResolver allows a Loader to control how a schema ref is resolved to a
// canonical name.
//
// When the generator follows a $ref from one schema file to another, it
// resolves the referenced file name to a canonical name. The canonical name
// identifies the loaded schema, and is used as the parent name when resolving
// references made by the referenced schema in turn.
//
// By default, references are resolved against the native filesystem with
// QualifiedFileName. Loaders that read schemas from somewhere other than the
// native filesystem should implement RefResolver to resolve references in
// their own namespace instead.
type RefResolver interface {
	// ResolveRef returns the canonical name of the fileName schema.
	//
	// parentFileName, if not empty, is the canonical name of the schema
	// containing the reference.
	ResolveRef(fileName, parentFileName string) (string, error)
}

// ResolveRef resolves fileName, referenced from the parentFileName schema,
// to a canonical schema name.
//
// If loader (or a Loader it wraps, following any Unwrap methods) implements
// RefResolver, resolution is delegated to it. Otherwise, fileName is resolved
// against the native filesystem with QualifiedFileName using resolveExtensions.
func ResolveRef(loader Loader, fileName, parentFileName string, resolveExtensions []string) (string, error) {
	for l := loader; l != nil; {
		if resolver, ok := l.(RefResolver); ok {
			qualified, err := resolver.ResolveRef(fileName, parentFileName)
			if err != nil {
				return "", fmt.Errorf("failed to resolve schema reference %q: %w", fileName, err)
			}

			return qualified, nil
		}

		unwrapper, ok := l.(interface{ Unwrap() Loader })
		if !ok {
			break
		}

		l = unwrapper.Unwrap()
	}

	return QualifiedFileName(fileName, parentFileName, resolveExtensions)
}

func NewCachedLoader(loader Loader, cache map[string]*Schema) *CachedLoader {
	return &CachedLoader{
		loader: loader,
		cache:  cache,
	}
}

type CachedLoader struct {
	loader Loader
	cache  map[string]*Schema
}

func (l *CachedLoader) Load(uri, parentURI string) (*Schema, error) {
	if schema, ok := l.cache[uri]; ok {
		return schema, nil
	}

	schema, err := l.loader.Load(uri, parentURI)
	if err != nil {
		return nil, errors.Join(ErrCannotLoadSchema, err)
	}

	l.cache[uri] = schema

	return schema, nil
}

// Unwrap returns the Loader wrapped by l.
func (l *CachedLoader) Unwrap() Loader {
	return l.loader
}

func NewFileLoader(resolveExtensions, yamlExtensions []string) *FileLoader {
	return &FileLoader{
		resolveExtensions: resolveExtensions,
		yamlExtensions:    toExtensionSet(yamlExtensions),
	}
}

type FileLoader struct {
	resolveExtensions []string
	yamlExtensions    map[string]bool
}

func (l *FileLoader) Load(fileName, parentFileName string) (*Schema, error) {
	qualified, err := QualifiedFileName(fileName, parentFileName, l.resolveExtensions)
	if err != nil {
		return nil, err
	}

	schema, err := l.parseFile(qualified)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func (l *FileLoader) parseFile(fileName string) (*Schema, error) {
	if l.yamlExtensions[path.Ext(fileName)] {
		sc, err := FromYAMLFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("error parsing YAML file %s: %w", fileName, err)
		}

		return sc, nil
	}

	sc, err := FromJSONFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON file %s: %w", fileName, err)
	}

	return sc, nil
}

func NewDefaultCacheLoader(resolveExtensions, yamlExtensions []string) *CachedLoader {
	return NewCachedLoader(NewDefaultMultiLoader(resolveExtensions, yamlExtensions), map[string]*Schema{})
}

func NewDefaultMultiLoader(resolveExtensions, yamlExtensions []string) MultiLoader {
	httpLoader := NewHTTPLoader(yamlExtensions)

	return MultiLoader{
		RefTypeFile:  NewFileLoader(resolveExtensions, yamlExtensions),
		RefTypeHTTP:  httpLoader,
		RefTypeHTTPS: httpLoader,
	}
}

type MultiLoader map[RefType]Loader

func (l MultiLoader) Load(uri, parentURI string) (*Schema, error) {
	ref, err := GetRefType(uri)
	if err != nil {
		return nil, err
	}

	loader, ok := l[ref]
	if !ok {
		return nil, ErrUnsupportedRefFormat
	}

	schema, err := loader.Load(uri, parentURI)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema %q: %w", uri, err)
	}

	return schema, nil
}

func NewHTTPLoader(yamlExtensions []string) *HTTPLoader {
	return &HTTPLoader{YAMLExtensions: toExtensionSet(yamlExtensions)}
}

type HTTPLoader struct {
	YAMLExtensions map[string]bool
}

func (l *HTTPLoader) Load(uri, parentURI string) (*Schema, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	if u.Scheme == "http" || u.Scheme == "https" {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to perform request: %w", err)
		}

		defer func() {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
		}()

		switch resp.Header.Get("Content-Type") {
		case "application/json":
			return FromJSONReader(resp.Body)

		case "application/yaml", "application/x-yaml", "text/yaml", "text/x-yaml":
			return FromYAMLReader(resp.Body)

		default:
			if l.YAMLExtensions[path.Ext(u.Path)] {
				return FromYAMLReader(resp.Body)
			}

			return FromJSONReader(resp.Body)
		}
	}

	return nil, fmt.Errorf("%w: %q", ErrUnsupportedURL, uri)
}

func QualifiedFileName(fileName, parentFileName string, resolveExtensions []string) (string, error) {
	r, err := GetRefType(fileName)
	if err != nil {
		return "", err
	}

	if r != RefTypeFile {
		return fileName[strings.Index(fileName, "://")+3:], nil
	}

	fileName = strings.TrimPrefix(fileName, "file://")

	if !filepath.IsAbs(fileName) {
		fileName = filepath.Join(filepath.Dir(parentFileName), fileName)
	}

	exts := append([]string{""}, resolveExtensions...)
	for _, ext := range exts {
		qualified := fileName + ext

		if !fileExists(qualified) {
			continue
		}

		var err error

		qualified, err = filepath.EvalSymlinks(qualified)
		if err != nil {
			return "", fmt.Errorf("error resolving symlinks in %s: %w", qualified, err)
		}

		return qualified, nil
	}

	return "", fmt.Errorf("%w %q", ErrCannotResolveSchema, fileName)
}

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)

	return err == nil || !os.IsNotExist(err)
}

func toExtensionSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))

	for _, item := range items {
		if !strings.HasPrefix(item, ".") {
			item = "." + item
		}

		set[item] = true
	}

	return set
}
