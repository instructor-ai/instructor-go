package instructor

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

func toPtr[T any](val T) *T {
	return &val
}

func prepend[T any](to []T, from T) []T {
	return append([]T{from}, to...)
}

func fetchMediaType(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return "", fmt.Errorf("could not determine content type")
	}

	return contentType, nil
}

func urlToBase64(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
