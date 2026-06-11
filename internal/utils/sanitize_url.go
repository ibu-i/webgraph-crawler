package utils

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"
)

func SanitizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	u.RawQuery = ""
	u.Fragment = ""

	host := u.Hostname()
	port := u.Port()

	if (u.Scheme == "http" && port == "80") ||
		(u.Scheme == "https" && port == "443") {
		u.Host = host
	}

	cleanPath := path.Clean(u.Path)

	if cleanPath == "." {
		cleanPath = "/"
	}

	switch cleanPath {
	case "/index.html",
		"/index.htm",
		"/index.php",
		"/default.aspx":
		cleanPath = "/"
	}

	u.Path = cleanPath

	if u.Path != "/" {
		u.Path = strings.TrimRight(u.Path, "/")
	}

	if net.ParseIP(host) == nil && host == "" {
		return "", fmt.Errorf("invalid host")
	}

	return u.String(), nil
}
