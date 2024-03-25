package httputil

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/xerrors"
)

type THeader struct {
	Key   string
	Value string
}

func ParseRequest(ctx context.Context, reqStr string, body io.Reader) (*http.Request, error) {
	var method, scheme, authority, path string
	var headers []*THeader
	lines := strings.Split(reqStr, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, ": ")
		key := parts[0]
		value := parts[1]
		if key[0] == ':' {
			switch key[1:] {
			case "method":
				method = value
			case "scheme":
				scheme = value
			case "authority":
				authority = value
			case "path":
				path = value
			}
		} else {
			headers = append(headers, &THeader{
				Key:   key,
				Value: value,
			})
		}
	}
	u, err := url.Parse(path)
	if err != nil {
		return nil, xerrors.Errorf("unable to parse path %v: %w", path, err)
	}
	u.Scheme = scheme
	u.Host = authority
	req, err := http.NewRequestWithContext(ctx, method, "", body)
	if err != nil {
		return nil, xerrors.Errorf("unable to initialize request: %w", err)
	}
	req.URL = u
	for _, header := range headers {
		req.Header.Add(header.Key, header.Value)
	}
	return req, nil
}
