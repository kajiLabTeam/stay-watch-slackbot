package lib

import (
	"net/http"

	"github.com/kajiLabTeam/stay-watch-slackbot/config"
)

// SharedHTTPClient は全てのHTTPリクエストで共有されるクライアント
var SharedHTTPClient *http.Client

func init() {
	SharedHTTPClient = &http.Client{
		Timeout: config.HTTPClient.Timeout,
	}
}
