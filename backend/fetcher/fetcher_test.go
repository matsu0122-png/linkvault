package fetcher

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestParseMetadata(t *testing.T) {
	base, err := url.Parse("https://example.com/articles/go")
	if err != nil {
		t.Fatalf("invalid base url: %v", err)
	}

	t.Run("title/description/OGP画像/faviconを取得できる", func(t *testing.T) {
		html := `<html><head>
			<title>Go Documentation</title>
			<meta name="description" content="fallback description">
			<meta property="og:description" content="OGP description">
			<meta property="og:image" content="/images/og.png">
			<link rel="icon" href="/favicon.png">
		</head><body></body></html>`

		got := parseMetadata(strings.NewReader(html), base)

		if got.Title != "Go Documentation" {
			t.Errorf("Title = %q", got.Title)
		}
		if got.Description != "OGP description" {
			t.Errorf("Description = %q, want og:description to take priority", got.Description)
		}
		if got.ImageURL != "https://example.com/images/og.png" {
			t.Errorf("ImageURL = %q", got.ImageURL)
		}
		if got.FaviconURL != "https://example.com/favicon.png" {
			t.Errorf("FaviconURL = %q", got.FaviconURL)
		}
	})

	t.Run("og:descriptionが無ければmeta descriptionを使う", func(t *testing.T) {
		html := `<html><head><meta name="description" content="fallback description"></head></html>`

		got := parseMetadata(strings.NewReader(html), base)

		if got.Description != "fallback description" {
			t.Errorf("Description = %q", got.Description)
		}
	})

	t.Run("headタグの外にあるtitleは無視して既に見つけたものを優先する", func(t *testing.T) {
		html := `<html><head><title>Go Documentation</title></head><body><title>ignored</title></body></html>`

		got := parseMetadata(strings.NewReader(html), base)

		if got.Title != "Go Documentation" {
			t.Errorf("Title = %q", got.Title)
		}
	})

	t.Run("何も無ければ全て空文字", func(t *testing.T) {
		html := `<html><head></head><body>no metadata here</body></html>`

		got := parseMetadata(strings.NewReader(html), base)

		if got.Title != "" || got.Description != "" || got.ImageURL != "" || got.FaviconURL != "" {
			t.Errorf("expected all fields empty, got %+v", got)
		}
	})

	t.Run("http/https以外のスキームのimage/faviconは採用しない", func(t *testing.T) {
		html := `<html><head>
			<meta property="og:image" content="javascript:alert(1)">
			<link rel="icon" href="data:image/png;base64,AAAA">
		</head></html>`

		got := parseMetadata(strings.NewReader(html), base)

		if got.ImageURL != "" || got.FaviconURL != "" {
			t.Errorf("expected unsafe schemes to be dropped, got %+v", got)
		}
	})
}

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		ip      string
		blocked bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"10.0.0.5", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true}, // cloud metadata endpoint
		{"0.0.0.0", true},
		{"8.8.8.8", false},
		{"93.184.216.34", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("invalid test IP: %s", tt.ip)
			}
			if got := isBlockedIP(ip); got != tt.blocked {
				t.Errorf("isBlockedIP(%s) = %v, want %v", tt.ip, got, tt.blocked)
			}
		})
	}
}

func TestFetchMetadata(t *testing.T) {
	t.Run("正常なページからmetadataを取得できる", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html><head><title>Example</title><meta property="og:image" content="/og.png"></head></html>`))
		}))
		defer ts.Close()

		f := New()
		f.client.Transport = http.DefaultTransport // httptest binds to 127.0.0.1; bypass the SSRF guard just for this happy-path test

		got, err := f.FetchMetadata(ts.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Title != "Example" {
			t.Errorf("Title = %q", got.Title)
		}
		if got.ImageURL != ts.URL+"/og.png" {
			t.Errorf("ImageURL = %q, want %q", got.ImageURL, ts.URL+"/og.png")
		}
	})

	t.Run("loopbackアドレスへのアクセスはブロックされる", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html><head><title>Example</title></head></html>`))
		}))
		defer ts.Close()

		f := New()

		if _, err := f.FetchMetadata(ts.URL); err == nil {
			t.Fatal("expected loopback access to be blocked, got nil error")
		}
	})

	t.Run("HTTPステータスが2xx以外ならエラー", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		f := New()
		f.client.Transport = http.DefaultTransport

		if _, err := f.FetchMetadata(ts.URL); err == nil {
			t.Fatal("expected error for non-2xx status, got nil")
		}
	})

	t.Run("サポートされないスキームはエラー", func(t *testing.T) {
		f := New()

		if _, err := f.FetchMetadata("file:///etc/passwd"); err == nil {
			t.Fatal("expected error for unsupported scheme, got nil")
		}
	})
}
