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

func NewFileLoader(resolveExtensions, yamlExtensions []string) *FileLoader {
	return &FileLoader{
		resolveExtensions: resolveExtensions,
		yamlExtensions:    toSet(yamlExtensions),
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

	return loader.Load(uri, parentURI)
}

func NewHTTPLoader(yamlExtensions []string) *HTTPLoader {
	return &HTTPLoader{YAMLExtensions: toSet(yamlExtensions)}
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

func toSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}

	return set
}
