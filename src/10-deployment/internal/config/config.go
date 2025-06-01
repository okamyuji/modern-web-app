package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// アプリケーション設定
	AppName    string
	Port       string
	Env        string
	Version    string
	BuildTime  string
	DebugMode  bool

	// サーバー設定
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	// データベース設定
	DatabaseURL     string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration

	// セキュリティ設定
	SessionSecret  string
	CSRFSecret     string
	AllowedOrigins []string
	TLSEnabled     bool
	TLSCertFile    string
	TLSKeyFile     string

	// Redis設定
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// SMTP設定
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string

	// 監視・ログ設定
	LogLevel          string
	SentryDSN         string
	PrometheusEnabled bool
	PrometheusPort    string

	// 外部サービス設定
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string

	// AWS設定
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSS3Bucket        string

	// その他
	EnableSwagger bool
	APIRateLimit  int
	CORSEnabled   bool
}

func Load() (*Config, error) {
	// 環境に応じて.envファイルを読み込み
	env := getEnv("ENV", "development")
	
	envFiles := []string{
		fmt.Sprintf(".env.%s.local", env),
		fmt.Sprintf(".env.%s", env),
		".env.local",
		".env",
	}

	// 利用可能な環境ファイルを順番に読み込み
	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err == nil {
			fmt.Printf("Loaded config from: %s\n", envFile)
			break
		}
	}

	config := &Config{
		// アプリケーション設定
		AppName:   getEnv("APP_NAME", "deployment-demo"),
		Port:      getEnv("PORT", "8080"),
		Env:       env,
		Version:   getEnv("VERSION", "dev"),
		BuildTime: getEnv("BUILD_TIME", "unknown"),
		DebugMode: env == "development",

		// サーバー設定
		ReadTimeout:     getDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getDuration("WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:     getDuration("IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 30*time.Second),

		// データベース設定
		DatabaseURL:     getEnv("DATABASE_URL", "sqlite://deployment_demo.db"),
		MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),

		// セキュリティ設定
		SessionSecret:  getEnvRequired("SESSION_SECRET", "SESSION_SECRET is required"),
		CSRFSecret:     getEnvRequired("CSRF_SECRET", "CSRF_SECRET is required"),
		AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"http://localhost:8080"}),
		TLSEnabled:     getBool("TLS_ENABLED", false),
		TLSCertFile:    getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:     getEnv("TLS_KEY_FILE", ""),

		// Redis設定
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getInt("REDIS_DB", 0),

		// SMTP設定
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),

		// 監視・ログ設定
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		SentryDSN:         getEnv("SENTRY_DSN", ""),
		PrometheusEnabled: getBool("PROMETHEUS_ENABLED", true),
		PrometheusPort:    getEnv("PROMETHEUS_PORT", "9090"),

		// 外部サービス設定
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),

		// AWS設定
		AWSRegion:          getEnv("AWS_REGION", "ap-northeast-1"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSS3Bucket:        getEnv("AWS_S3_BUCKET", ""),

		// その他
		EnableSwagger: getBool("ENABLE_SWAGGER", env == "development"),
		APIRateLimit:  getInt("API_RATE_LIMIT", 100),
		CORSEnabled:   getBool("CORS_ENABLED", true),
	}

	// 設定の検証
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}


func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) IsTesting() bool {
	return c.Env == "test" || c.Env == "testing"
}

func (c *Config) GetDatabaseDriver() string {
	if strings.HasPrefix(c.DatabaseURL, "postgres://") {
		return "postgres"
	}
	return "sqlite3"
}

func (c *Config) Validate() error {
	var errors []string

	// 本番環境での必須チェック
	if c.IsProduction() {
		if c.SessionSecret == "change-this-in-production-please-use-random-64-chars" {
			errors = append(errors, "SESSION_SECRET must be changed in production")
		}
		if c.CSRFSecret == "change-this-csrf-secret-in-production-use-random-32-chars" {
			errors = append(errors, "CSRF_SECRET must be changed in production")
		}
		if c.LogLevel == "debug" {
			errors = append(errors, "LOG_LEVEL should not be 'debug' in production")
		}
		if c.DebugMode {
			errors = append(errors, "debug mode should be disabled in production")
		}

		// TLS証明書の確認
		if c.TLSEnabled && (c.TLSCertFile == "" || c.TLSKeyFile == "") {
			errors = append(errors, "TLS_CERT_FILE and TLS_KEY_FILE are required when TLS is enabled")
		}
	}

	// ポート番号の妥当性確認
	if port, err := strconv.Atoi(c.Port); err != nil || port < 1 || port > 65535 {
		errors = append(errors, "invalid PORT number")
	}

	// データベースURL形式確認
	if c.DatabaseURL == "" {
		errors = append(errors, "DATABASE_URL is required")
	}

	// セキュリティ設定の確認
	if len(c.SessionSecret) < 32 {
		errors = append(errors, "SESSION_SECRET must be at least 32 characters")
	}
	if len(c.CSRFSecret) < 16 {
		errors = append(errors, "CSRF_SECRET must be at least 16 characters")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{AppName: %s, Env: %s, Version: %s, Port: %s}", 
		c.AppName, c.Env, c.Version, c.Port)
}

// ヘルパー関数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string, _ string) string {
	value := os.Getenv(key)
	if value == "" {
		// デフォルト値を開発環境用に設定
		switch key {
		case "SESSION_SECRET":
			return "change-this-in-production-please-use-random-64-chars"
		case "CSRF_SECRET":
			return "change-this-csrf-secret-in-production-use-random-32-chars"
		default:
			return ""
		}
	}
	return value
}

func getInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}