package lib

import (
	"net/http"
	"time"
)

// SharedHTTPClient は全てのHTTPリクエストで共有されるクライアント
var SharedHTTPClient *http.Client

func init() {
	SharedHTTPClient = &http.Client{
		Timeout: 30 * time.Second,
	}
}
