package instructor

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// Removes any prefixes before the JSON (like "Sure, here you go:")
func trimPrefixBeforeJSON(jsonStr string) string {
	startObject := strings.IndexByte(jsonStr, '{')
	startArray := strings.IndexByte(jsonStr, '[')

	var start int
	if startObject == -1 && startArray == -1 {
		return jsonStr // No opening brace or bracket found, return the original string
	} else if startObject == -1 {
		start = startArray
	} else if startArray == -1 {
		start = startObject
	} else {
		start = min(startObject, startArray)
	}

	return jsonStr[start:]
}

// Removes any postfixes after the JSON
func trimPostfixAfterJSON(jsonStr string) string {
	endObject := strings.LastIndexByte(jsonStr, '}')
	endArray := strings.LastIndexByte(jsonStr, ']')

	var end int
	if endObject == -1 && endArray == -1 {
		return jsonStr // No closing brace or bracket found, return the original string
	} else if endObject == -1 {
		end = endArray
	} else if endArray == -1 {
		end = endObject
	} else {
		end = max(endObject, endArray)
	}

	return jsonStr[:end+1]
}

// Extracts the JSON by trimming prefixes and postfixes
func extractJSON(jsonStr string) string {
	trimmedPrefix := trimPrefixBeforeJSON(jsonStr)
	trimmedJSON := trimPostfixAfterJSON(trimmedPrefix)
	return trimmedJSON
}
