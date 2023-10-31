package mail

import (
	"errors"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
)

var urlR = regexp.MustCompile(`https?://[^\s]+|ftp://[^\s]+`)

func urls(data string, ignores ...string) []*url.URL {
	out := []*url.URL{}

	var dedup = make(map[string]bool)
	allU := urlR.FindAllString(data, -1)
uloop:
	for _, u := range allU {
		u = strings.TrimSuffix(u, ">")
		u = strings.TrimSuffix(u, "\"")

		if dedup[u] {
			continue
		}

		dedup[u] = true

		parsed, err := url.Parse(u)
		if err != nil {
			slog.Warn(
				"Error parsing URL",
				"error", err.Error(),
				"url", u,
			)
		}

		if parsed == nil {
			continue
		}

		parsed, err = stripSafeLN(parsed)
		if err != nil {
			slog.Warn(
				"Error stripping safe link",
				"error", err.Error(),
				"url", u,
			)
		}

		// check for duplicate
		if dedup[parsed.String()] {
			continue
		}

		for _, ignore := range ignores {
			if strings.HasSuffix(parsed.Host, ignore) {
				continue uloop
			}
		}

		out = append(out, parsed)
	}

	return out
}

var ErrEmpty = errors.New("empty safe link")

//nolint:gochecknoglobals // this is a list of known safe link suffixes
var knownSafeLinkSuffix = []string{
	"safelinks.protection.outlook.com",
}

func stripSafeLN(ln *url.URL) (*url.URL, error) {
	if ln == nil {
		return nil, ErrEmpty
	}

	if !isSafeLn(ln) {
		return ln, nil
	}

	u, ok := ln.Query()["url"]
	if !ok || len(u) == 0 {
		return nil, ErrEmpty
	}

	return url.Parse(u[0])
}

func isSafeLn(ln *url.URL) bool {
	for _, s := range knownSafeLinkSuffix {
		if strings.HasSuffix(ln.Host, s) {
			return true
		}
	}

	return false
}
