// Package fetcher retrieves the <title> of a web page so links can be saved
// without the user having to type a title by hand.
package fetcher

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html"
)

const (
	requestTimeout = 5 * time.Second
	dialTimeout    = 3 * time.Second
	maxRedirects   = 3
	maxBodyBytes   = 512 * 1024 // 512KB, well beyond where <title> normally appears
)

var errTitleNotFound = errors.New("title not found")

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

// FetchTitle fetches rawURL and returns the text content of its <title> element.
func (f *Fetcher) FetchTitle(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme: %s", parsed.Scheme)
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return parseTitle(io.LimitReader(resp.Body, maxBodyBytes))
}

// parseTitle scans HTML and returns the text of the first <title> element.
func parseTitle(r io.Reader) (string, error) {
	tokenizer := html.NewTokenizer(r)
	inTitle := false

	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			return "", errTitleNotFound
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			inTitle = string(name) == "title"
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			if string(name) == "title" {
				inTitle = false
			}
		case html.TextToken:
			if inTitle {
				title := strings.TrimSpace(string(tokenizer.Text()))
				if title != "" {
					return title, nil
				}
			}
		}
	}
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
