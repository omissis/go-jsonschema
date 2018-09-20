package schemas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func FromFile(fileName string) (*Schema, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return FromReader(f)
}

func FromReader(r io.Reader) (*Schema, error) {
	var schema Schema
	if err := json.NewDecoder(r).Decode(&schema); err != nil {
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
