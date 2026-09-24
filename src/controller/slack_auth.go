package controller

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

// VerifySlackSignature は X-Slack-Signature を検証するミドルウェア。
// スラッシュコマンドのエンドポイントは SlashCommandParse が署名検証を行わないため、
// これを経由させて Slack 以外からのリクエストを拒否する。
// 検証後、後続の SlashCommandParse がボディを読めるよう Request.Body を復元する。
func VerifySlackSignature() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()
		if err != nil {
			respondError(c, http.StatusBadRequest, "bad request")
			c.Abort()
			return
		}

		sv, err := slack.NewSecretsVerifier(c.Request.Header, signingSecret)
		if err != nil {
			respondError(c, http.StatusBadRequest, "bad request")
			c.Abort()
			return
		}
		if _, err := sv.Write(body); err != nil {
			respondError(c, http.StatusInternalServerError, msgInternalServerError)
			c.Abort()
			return
		}
		if err := sv.Ensure(); err != nil {
			respondError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Next()
	}
}
