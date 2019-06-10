package schemas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/atombender/go-jsonschema/pkg/yamlutils"
)

func FromJSONFile(fileName string) (*Schema, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return FromJSONReader(f)
}

func FromJSONReader(r io.Reader) (*Schema, error) {
	var schema Schema
	if err := json.NewDecoder(r).Decode(&schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

func FromYAMLFile(fileName string) (*Schema, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return FromYAMLReader(f)
}

func FromYAMLReader(r io.Reader) (*Schema, error) {
	// Marshal to JSON first because YAML decoder doesn't understand JSON tags
	var m map[string]interface{}
	if err := yaml.NewDecoder(r).Decode(&m); err != nil {
		return nil, err
	}
	yamlutils.FixMapKeys(m)

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var schema Schema
	if err = json.Unmarshal(b, &schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

type Loader struct {
	workingDir string
}

func (l *Loader) Load(fromURL string) (io.ReadCloser, error) {
	u, err := url.Parse(fromURL)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "http" || u.Scheme == "https" {
		resp, err := http.Get(fromURL)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	}

	if (u.Scheme == "" || u.Scheme == "file") && u.Host == "" && u.Path != "" {
		return os.Open(filepath.Join(l.workingDir, u.Path))
	}

	return nil, fmt.Errorf("schema reference must a file name or HTTP URL: %q", fromURL)
}
