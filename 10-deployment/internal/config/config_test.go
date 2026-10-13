package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func setProductionEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ENV", "production")
	t.Setenv("SESSION_SECRET", strings.Repeat("s", 64))
	t.Setenv("CSRF_SECRET", strings.Repeat("c", 32))
}

func TestLoad_RequiresAPITokenInProduction(t *testing.T) {
	setProductionEnv(t)
	t.Setenv("API_TOKEN", "")

	_, err := Load()

	if err == nil || !strings.Contains(err.Error(), "API_TOKEN") {
		t.Fatalf("err = %v, want an API_TOKEN validation error", err)
	}
}

func TestLoad_APITokenLengthInProduction(t *testing.T) {
	cases := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{"one below minimum", minAPITokenLength - 1, true},
		{"at minimum", minAPITokenLength, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setProductionEnv(t)
			t.Setenv("API_TOKEN", strings.Repeat("t", tc.length))

			cfg, err := Load()

			if tc.wantErr != (err != nil) {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && cfg.APIToken != strings.Repeat("t", tc.length) {
				t.Fatal("APIToken not loaded from API_TOKEN")
			}
		})
	}
}

func TestLoad_RejectsNonPositiveAPIRateLimit(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("API_RATE_LIMIT", "0")

	_, err := Load()

	if err == nil || !strings.Contains(err.Error(), "API_RATE_LIMIT") {
		t.Fatalf("err = %v, want an API_RATE_LIMIT validation error", err)
	}
}

func TestLoad_AllowsMissingAPITokenOutsideProduction(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("API_TOKEN", "")

	cfg, err := Load()

	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.APIToken != "" {
		t.Fatalf("APIToken = %q, want empty", cfg.APIToken)
	}
}

func validProductionConfig() Config {
	return Config{
		Env:           "production",
		Port:          "8080",
		DatabaseURL:   "sqlite://app.db",
		SessionSecret: strings.Repeat("s", 32),
		CSRFSecret:    strings.Repeat("c", 16),
		APIToken:      strings.Repeat("t", minAPITokenLength),
		APIRateLimit:  1,
		LogLevel:      "info",
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{"valid production config", func(*Config) {}, ""},
		{"default session secret in production", func(c *Config) {
			c.SessionSecret = "change-this-in-production-please-use-random-64-chars"
		}, "SESSION_SECRET must be changed"},
		{"default csrf secret in production", func(c *Config) {
			c.CSRFSecret = "change-this-csrf-secret-in-production-use-random-32-chars"
		}, "CSRF_SECRET must be changed"},
		{"debug log level in production", func(c *Config) { c.LogLevel = "debug" }, "LOG_LEVEL"},
		{"debug log level outside production", func(c *Config) { c.Env = "development"; c.LogLevel = "debug" }, ""},
		{"debug mode in production", func(c *Config) { c.DebugMode = true }, "debug mode"},
		{"tls without cert", func(c *Config) { c.TLSEnabled = true; c.TLSKeyFile = "k" }, "TLS_CERT_FILE"},
		{"tls without key", func(c *Config) { c.TLSEnabled = true; c.TLSCertFile = "c" }, "TLS_CERT_FILE"},
		{"tls with cert and key", func(c *Config) { c.TLSEnabled = true; c.TLSCertFile = "c"; c.TLSKeyFile = "k" }, ""},
		{"tls without files outside production", func(c *Config) { c.Env = "development"; c.TLSEnabled = true }, ""},
		{"non-numeric port", func(c *Config) { c.Port = "abc" }, "invalid PORT"},
		{"port 0", func(c *Config) { c.Port = "0" }, "invalid PORT"},
		{"port 1", func(c *Config) { c.Port = "1" }, ""},
		{"port 65535", func(c *Config) { c.Port = "65535" }, ""},
		{"port 65536", func(c *Config) { c.Port = "65536" }, "invalid PORT"},
		{"empty database url", func(c *Config) { c.DatabaseURL = "" }, "DATABASE_URL"},
		{"rate limit 0", func(c *Config) { c.APIRateLimit = 0 }, "API_RATE_LIMIT"},
		{"session secret 31 chars", func(c *Config) { c.SessionSecret = strings.Repeat("s", 31) }, "SESSION_SECRET must be at least"},
		{"csrf secret 15 chars", func(c *Config) { c.CSRFSecret = strings.Repeat("c", 15) }, "CSRF_SECRET must be at least"},
		{"short api token outside production", func(c *Config) { c.Env = "development"; c.APIToken = "" }, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validProductionConfig()
			tc.mutate(&cfg)

			err := cfg.Validate()

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tc.wantErr)
			}
		})
	}
}

func TestLoad_APITokenRequirementByEnvironment(t *testing.T) {
	cases := []struct {
		env     string
		token   string
		wantErr bool
	}{
		{"development", "", false},
		{"test", "", false},
		{"testing", "", false},
		{"staging", "", true},
		{"prodution", "", true},
		{"staging", strings.Repeat("t", minAPITokenLength), false},
		{"staging", strings.Repeat("t", minAPITokenLength-1), true},
	}
	for _, tc := range cases {
		t.Run(tc.env+"/"+strconv.Itoa(len(tc.token)), func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("ENV", tc.env)
			t.Setenv("API_TOKEN", tc.token)

			_, err := Load()

			if tc.wantErr != (err != nil) {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr && !strings.Contains(err.Error(), "API_TOKEN") {
				t.Fatalf("err = %v, want an API_TOKEN validation error", err)
			}
		})
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	os.Unsetenv(key)
}

func TestLoad_ReadsENVFromDotEnvFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("ENV=production\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	unsetEnv(t, "ENV")
	t.Setenv("API_TOKEN", "")

	_, err := Load()

	if err == nil || !strings.Contains(err.Error(), "API_TOKEN") {
		t.Fatalf("err = %v, want production validation from ENV in .env", err)
	}
}

func TestLoad_DefaultsToProductionWhenENVMissing(t *testing.T) {
	for name, set := range map[string]func(*testing.T){
		"unset": func(t *testing.T) { unsetEnv(t, "ENV") },
		"empty": func(t *testing.T) { t.Setenv("ENV", "") },
	} {
		t.Run(name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			set(t)
			t.Setenv("API_TOKEN", "")

			_, err := Load()

			if err == nil || !strings.Contains(err.Error(), "API_TOKEN") {
				t.Fatalf("err = %v, want validation as production", err)
			}
		})
	}
}

func TestLoad_DotEnvCanSelectDevelopment(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("ENV=development\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	unsetEnv(t, "ENV")
	t.Setenv("API_TOKEN", "")

	cfg, err := Load()

	if err != nil || cfg.Env != "development" {
		t.Fatalf("cfg.Env = %q, err = %v, want development without error", cfg.Env, err)
	}
}
