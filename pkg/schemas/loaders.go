package schemas

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
		ResolveExtensions: resolveExtensions,
		YAMLExtensions:    yamlExtensions,
	}
}

type FileLoader struct {
	ResolveExtensions []string
	YAMLExtensions    []string
}

func (l *FileLoader) Load(fileName, parentFileName string) (*Schema, error) {
	qualified, err := QualifiedFileName(fileName, parentFileName, l.ResolveExtensions)
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
	for _, yamlExt := range l.YAMLExtensions {
		if strings.HasSuffix(fileName, yamlExt) {
			sc, err := FromYAMLFile(fileName)
			if err != nil {
				return nil, fmt.Errorf("error parsing YAML file %s: %w", fileName, err)
			}

			return sc, nil
		}
	}

	sc, err := FromJSONFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON file %s: %w", fileName, err)
	}

	return sc, nil
}

func NewMultiLoader(workingDir string) *MultiLoader {
	return &MultiLoader{
		workingDir: workingDir,
	}
}

type MultiLoader struct {
	workingDir string
}

func (l *MultiLoader) Load(uri, parentURI string) (*Schema, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	if u.Scheme == "http" || u.Scheme == "https" {
		resp, err := http.NewRequestWithContext(context.Background(), http.MethodGet, uri, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		defer resp.Body.Close()

		switch resp.Header.Get("Content-Type") {
		case "application/json":
			return FromJSONReader(resp.Body)

		case "application/yaml", "application/x-yaml", "text/yaml", "text/x-yaml":
			return FromYAMLReader(resp.Body)

		default:
			return nil, fmt.Errorf("%w: %q", ErrUnsupportedContentType, resp.Header.Get("Content-Type"))
		}
	}

	if (u.Scheme == "" || u.Scheme == "file") && u.Host == "" && u.Path != "" {
		rc, err := os.Open(filepath.Join(l.workingDir, u.Path))
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		defer rc.Close()

		switch filepath.Ext(u.Path) {
		case ".json":
			return FromYAMLReader(rc)

		case ".yaml", ".yml":
			return FromJSONReader(rc)

		default:
			return nil, fmt.Errorf("%w: %q", ErrUnsupportedFileExtension, filepath.Ext(u.Path))
		}
	}

	return nil, fmt.Errorf("%w: %q", ErrUnsupportedURL, uri)
}

func QualifiedFileName(fileName, parentFileName string, resolveExtensions []string) (string, error) {
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
