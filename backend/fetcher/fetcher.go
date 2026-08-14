// Package fetcher retrieves page metadata (title, description, OGP image,
// favicon) for a URL so links can be saved without the user typing them by
// hand.
package fetcher

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html"

	"github.com/matsu0122-png/linkvault/backend/model"
)

const (
	requestTimeout = 5 * time.Second
	dialTimeout    = 3 * time.Second
	maxRedirects   = 3
	maxBodyBytes   = 512 * 1024 // 512KB, well beyond where <head> normally ends
)

type Fetcher struct {
	client *http.Client
}

func New() *Fetcher {
	dialer := &net.Dialer{
		Timeout: dialTimeout,
		Control: blockUnsafeAddresses,
	}

	return &Fetcher{
		client: &http.Client{
			Timeout: requestTimeout,
			Transport: &http.Transport{
				DialContext: dialer.DialContext,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return fmt.Errorf("stopped after %d redirects", maxRedirects)
				}
				return nil
			},
		},
	}
}

// FetchMetadata fetches rawURL and extracts title, description, OGP image,
// and favicon from the response HTML. Fields that can't be found are left
// as empty strings rather than causing an error.
func (f *Fetcher) FetchMetadata(rawURL string) (model.Metadata, error) {
	if err := validateFetchURL(rawURL); err != nil {
		return model.Metadata{}, err
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return model.Metadata{}, fmt.Errorf("build request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return model.Metadata{}, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.Metadata{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return parseMetadata(io.LimitReader(resp.Body, maxBodyBytes), resp.Request.URL), nil
}

// CheckAlive reports whether rawURL currently responds successfully (2xx,
// after following redirects). A non-nil error means the check itself failed
// (invalid URL, SSRF block, timeout, connection error, or a non-2xx
// response) — the caller should treat that the same as "not alive".
func (f *Fetcher) CheckAlive(rawURL string) (bool, error) {
	if err := validateFetchURL(rawURL); err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return false, fmt.Errorf("build request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return true, nil
}

func validateFetchURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported scheme: %s", parsed.Scheme)
	}
	return nil
}

// parseMetadata scans HTML and extracts title, description, OGP image, and
// favicon. It stops once </head> is seen, since these elements only ever
// appear there. base is used to resolve relative image/favicon URLs.
func parseMetadata(r io.Reader, base *url.URL) model.Metadata {
	tokenizer := html.NewTokenizer(r)

	var meta model.Metadata
	var ogDescription string
	inTitle := false

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			tok := tokenizer.Token()
			switch tok.Data {
			case "title":
				inTitle = tt == html.StartTagToken
			case "meta":
				name, property, content := metaAttrs(tok)
				switch {
				case property == "og:description":
					ogDescription = content
				case name == "description" && meta.Description == "":
					meta.Description = content
				case property == "og:image" && meta.ImageURL == "":
					meta.ImageURL = resolveURL(base, content)
				}
			case "link":
				rel, href := linkAttrs(tok)
				if (rel == "icon" || rel == "shortcut icon") && meta.FaviconURL == "" {
					meta.FaviconURL = resolveURL(base, href)
				}
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "title":
				inTitle = false
			case "head":
				if ogDescription != "" {
					meta.Description = ogDescription
				}
				return meta
			}
		case html.TextToken:
			if inTitle && meta.Title == "" {
				if text := strings.TrimSpace(string(tokenizer.Text())); text != "" {
					meta.Title = text
				}
			}
		}
	}

	if ogDescription != "" {
		meta.Description = ogDescription
	}

	return meta
}

func metaAttrs(tok html.Token) (name, property, content string) {
	for _, a := range tok.Attr {
		switch a.Key {
		case "name":
			name = a.Val
		case "property":
			property = a.Val
		case "content":
			content = a.Val
		}
	}
	return name, property, content
}

func linkAttrs(tok html.Token) (rel, href string) {
	for _, a := range tok.Attr {
		switch a.Key {
		case "rel":
			rel = a.Val
		case "href":
			href = a.Val
		}
	}
	return rel, href
}

// resolveURL resolves ref (possibly relative) against base and only returns
// it when the result is an http(s) URL.
func resolveURL(base *url.URL, ref string) string {
	if ref == "" {
		return ""
	}

	parsed, err := url.Parse(ref)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(parsed)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}

	return resolved.String()
}

// blockUnsafeAddresses is a net.Dialer.Control hook that runs right before a
// connection is established, using the actual IP being dialed. Checking here
// (rather than resolving DNS ourselves beforehand) also covers redirects and
// is immune to DNS rebinding between validation and connect.
func blockUnsafeAddresses(network, address string, c syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("invalid address: %s", host)
	}

	if isBlockedIP(ip) {
		return fmt.Errorf("blocked address: %s", ip)
	}

	return nil
}

func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified()
}
