package config

import (
	"reflect"
	"testing"
)

func TestParseOrigins(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "単一オリジン",
			raw:  "http://localhost:5173",
			want: []string{"http://localhost:5173"},
		},
		{
			name: "カンマ区切りで複数オリジン",
			raw:  "http://localhost:5173,chrome-extension://abcdefgh",
			want: []string{"http://localhost:5173", "chrome-extension://abcdefgh"},
		},
		{
			name: "空白はトリムされる",
			raw:  "http://localhost:5173, chrome-extension://abcdefgh ",
			want: []string{"http://localhost:5173", "chrome-extension://abcdefgh"},
		},
		{
			name: "空文字の要素は除外される",
			raw:  "http://localhost:5173,,",
			want: []string{"http://localhost:5173"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOrigins(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseOrigins(%q) = %#v, want %#v", tt.raw, got, tt.want)
			}
		})
	}
}
