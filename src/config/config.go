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
	LabNetworkCIDR   string
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

// BoardThresholds はboard APIの段階化閾値を保持する
type BoardThresholds struct {
	LikelihoodHigh float64
	LikelihoodMid  float64
	LikelihoodMin  float64
	ArrivalLikely  float64
	ArrivalMaybe   float64
}

// BoardTimeBlockConfig は1時間帯（昼/夕方/夜など）の設定を保持する
type BoardTimeBlockConfig struct {
	ID         string
	Label      string
	RangeLabel string
	StartMin   int
	EndMin     int
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
	TimeBlocks []BoardTimeBlockConfig
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
		LabNetworkCIDR: getEnv("CORS_LAB_NETWORK_CIDR", "192.168.100.0/23"),
		AllowMethods:   getEnvStringSlice("CORS_ALLOW_METHODS", []string{"GET", "POST"}),
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
		LikelihoodHigh: getEnvFloat("BOARD_LIKELIHOOD_HIGH_THRESHOLD", 0.5),
		LikelihoodMid:  getEnvFloat("BOARD_LIKELIHOOD_MID_THRESHOLD", 0.3),
		LikelihoodMin:  getEnvFloat("BOARD_LIKELIHOOD_MIN_THRESHOLD", 0.1),
		ArrivalLikely:  getEnvFloat("BOARD_ARRIVAL_LIKELY_THRESHOLD", 0.5),
		ArrivalMaybe:   getEnvFloat("BOARD_ARRIVAL_MAYBE_THRESHOLD", 0.3),
	}

	TimeBlocks = []BoardTimeBlockConfig{
		{
			ID:         "noon",
			Label:      "昼",
			RangeLabel: "12〜15時",
			StartMin:   getEnvInt("BOARD_NOON_START_MIN", 12*60),
			EndMin:     getEnvInt("BOARD_NOON_END_MIN", 15*60),
		},
		{
			ID:         "evening",
			Label:      "夕方",
			RangeLabel: "15〜19時",
			StartMin:   getEnvInt("BOARD_EVENING_START_MIN", 15*60),
			EndMin:     getEnvInt("BOARD_EVENING_END_MIN", 19*60),
		},
		{
			ID:         "night",
			Label:      "夜",
			RangeLabel: "19時〜",
			StartMin:   getEnvInt("BOARD_NIGHT_START_MIN", 19*60),
			EndMin:     getEnvInt("BOARD_NIGHT_END_MIN", 22*60),
		},
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
