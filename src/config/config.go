// Package config は環境変数から読み込むアプリケーション設定を集約する。
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return f
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return d
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}

// CORSConfig はCORSミドルウェアの設定を保持する
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// ServerConfig はHTTPサーバーの設定を保持する
type ServerConfig struct {
	Port           string
	TrustedProxies []string
}

// BoardThresholds はboard APIの来訪見込み閾値を保持する
type BoardThresholds struct {
	// ArrivalLikely はレスポンスには出さないが、閾値の再調整余地を残すため保持する
	ArrivalLikely float64
	// ArrivalMaybe は「今日来そうな人」の足切りに使う
	ArrivalMaybe float64
	// ActivityProbability は時間帯別の活動成立判定に使うGMM活動確率のしきい値
	ActivityProbability float64
}

// S3Config はS3互換オブジェクトストレージ（RustFS）の設定を保持する
type S3Config struct {
	Endpoint      string
	Region        string
	AccessKeyID   string
	SecretKey     string
	Bucket        string
	PublicBaseURL string
}

// Enabled はアップロードに必要な設定がそろっているかを返す
func (c S3Config) Enabled() bool {
	return c.Endpoint != "" && c.AccessKeyID != "" && c.SecretKey != "" && c.Bucket != ""
}

// DBConfig はDB接続の設定を保持する
type DBConfig struct {
	User          string
	Password      string
	Protocol      string
	DBName        string
	RetryCount    int
	RetryInterval time.Duration
}

// HTTPClientConfig は共有HTTPクライアントの設定を保持する
type HTTPClientConfig struct {
	Timeout time.Duration
}

var (
	CORS       CORSConfig
	Server     ServerConfig
	Board      BoardThresholds
	S3         S3Config
	DB         DBConfig
	HTTPClient HTTPClientConfig
)

func init() {
	CORS = CORSConfig{
		AllowOrigins: getEnvStringSlice("CORS_ALLOW_ORIGINS", []string{
			"https://staywatch.kajilab.net",
			"http://localhost:3000",
			"http://localhost:5173",
		}),
		AllowMethods: getEnvStringSlice("CORS_ALLOW_METHODS", []string{"GET", "POST"}),
		AllowHeaders: getEnvStringSlice("CORS_ALLOW_HEADERS", []string{
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Accept",
			"Authorization",
		}),
		AllowCredentials: getEnv("CORS_ALLOW_CREDENTIALS", "true") == "true",
		MaxAge:           getEnvDuration("CORS_MAX_AGE", 24*time.Hour),
	}

	Server = ServerConfig{
		Port:           getEnv("API_PORT", "8085"),
		TrustedProxies: getEnvStringSlice("TRUSTED_PROXIES", []string{"127.0.0.1"}),
	}

	Board = BoardThresholds{
		ArrivalLikely:       getEnvFloat("BOARD_ARRIVAL_LIKELY_THRESHOLD", 0.5),
		ArrivalMaybe:        getEnvFloat("BOARD_ARRIVAL_MAYBE_THRESHOLD", 0.3),
		ActivityProbability: getEnvFloat("BOARD_ACTIVITY_PROBABILITY_THRESHOLD", 0.3),
	}

	S3 = S3Config{
		Endpoint:      getEnv("S3_ENDPOINT", ""),
		Region:        getEnv("S3_REGION", "us-east-1"),
		AccessKeyID:   getEnv("S3_ACCESS_KEY_ID", ""),
		SecretKey:     getEnv("S3_SECRET_ACCESS_KEY", ""),
		Bucket:        getEnv("S3_BUCKET", ""),
		PublicBaseURL: getEnv("S3_PUBLIC_BASE_URL", ""),
	}

	DB = DBConfig{
		User:          getEnv("MYSQL_USER", ""),
		Password:      getEnv("MYSQL_PASSWORD", ""),
		Protocol:      getEnv("MYSQL_PROTOCOL", ""),
		DBName:        getEnv("MYSQL_DBNAME", ""),
		RetryCount:    getEnvInt("MYSQL_CONNECT_RETRY_COUNT", 10),
		RetryInterval: getEnvDuration("MYSQL_CONNECT_RETRY_INTERVAL", 2*time.Second),
	}

	HTTPClient = HTTPClientConfig{
		Timeout: getEnvDuration("HTTP_CLIENT_TIMEOUT", 30*time.Second),
	}
}
