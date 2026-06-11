package utils

import (
	"net/url"
	"strings"
)

func GetHost(rawURL string) (string, error) {
	sanitizedURL, err := SanitizeURL(rawURL)

	if err != nil {
		return "", err
	}

	parsedURL, err := url.Parse(sanitizedURL)
	if err != nil {
		return "", err
	}

	return strings.ToLower(parsedURL.Host), nil
}
