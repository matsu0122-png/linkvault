package fetcher

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseTitle(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		want    string
		wantErr bool
	}{
		{
			name: "titleがあれば取得できる",
			html: `<html><head><title>Go Documentation</title></head><body></body></html>`,
			want: "Go Documentation",
		},
		{
			name: "前後の空白はトリムされる",
			html: `<html><head><title>  Go Documentation  </title></head></html>`,
			want: "Go Documentation",
		},
		{
			name: "HTMLエンティティはデコードされる",
			html: `<html><head><title>Go &amp; Rust</title></head></html>`,
			want: "Go & Rust",
		},
		{
			name:    "titleが存在しなければエラー",
			html:    `<html><head></head><body>no title here</body></html>`,
			wantErr: true,
		},
		{
			name:    "titleが空文字なら見つからない扱い",
			html:    `<html><head><title></title></head></html>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTitle(strings.NewReader(tt.html))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got title %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
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

func TestFetchTitle(t *testing.T) {
	t.Run("正常なページからtitleを取得できる", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html><head><title>Example</title></head></html>`))
		}))
		defer ts.Close()

		f := New()
		f.client.Transport = http.DefaultTransport // httptest binds to 127.0.0.1; bypass the SSRF guard just for this happy-path test

		got, err := f.FetchTitle(ts.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "Example" {
			t.Errorf("got %q, want %q", got, "Example")
		}
	})

	t.Run("loopbackアドレスへのアクセスはブロックされる", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html><head><title>Example</title></head></html>`))
		}))
		defer ts.Close()

		f := New()

		if _, err := f.FetchTitle(ts.URL); err == nil {
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

		if _, err := f.FetchTitle(ts.URL); err == nil {
			t.Fatal("expected error for non-2xx status, got nil")
		}
	})

	t.Run("サポートされないスキームはエラー", func(t *testing.T) {
		f := New()

		if _, err := f.FetchTitle("file:///etc/passwd"); err == nil {
			t.Fatal("expected error for unsupported scheme, got nil")
		}
	})
}
