package config

import (
	"strings"
	"testing"
)

func TestValidateJWTSecret(t *testing.T) {
	valid32 := strings.Repeat("k", 32)

	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "empty", in: "", wantErr: true},
		{name: "whitespace only", in: "   \t  ", wantErr: true},
		{name: "31 bytes", in: strings.Repeat("k", 31), wantErr: true},
		{name: "exactly 32 bytes", in: valid32, want: valid32},
		{name: "surrounding whitespace trimmed", in: "\n " + valid32 + " \r\n", want: valid32},
		{name: "padded to 32 with spaces", in: " " + strings.Repeat("k", 30) + " ", wantErr: true},
		{name: "inner whitespace", in: "abc def" + strings.Repeat("k", 25), wantErr: true},
		{name: "control character", in: strings.Repeat("k", 31) + "\x7f", wantErr: true},
		{name: "multibyte counts as bytes", in: strings.Repeat("密钥", 6), want: strings.Repeat("密钥", 6)},
		{name: "default placeholder", in: "changemechangemechangemechangeme12", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateJWTSecret(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got secret %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLoadTrimsJWTSecret(t *testing.T) {
	secret := strings.Repeat("k", 32)
	t.Setenv("APP_POSTGRES_DSN", "postgres://localhost/craftshelf")
	t.Setenv("APP_JWT_SECRET", "  "+secret+"\n")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.JWTSecret != secret {
		t.Fatalf("JWTSecret = %q, want trimmed %q", cfg.JWTSecret, secret)
	}
}
