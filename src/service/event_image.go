package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"github.com/kajiLabTeam/stay-watch-slackbot/model"
)

// maxEventImageBytes は取り込む活動画像の上限サイズ
const maxEventImageBytes = 10 << 20 // 10MiB

// EventImageURL は保存済みキーから公開URLを組み立てる。
// キーが未登録、またはストレージが未設定の場合は nil を返す（レスポンスでは null になる）。
func EventImageURL(imageKey *string) *string {
	if imageKey == nil {
		return nil
	}
	url := lib.EventImages.PublicURL(*imageKey)
	if url == "" {
		return nil
	}
	return &url
}

// RegisterEventImageFromSlack は Slack にアップロードされた画像を取得して
// オブジェクトストレージへ保存し、events.image_key を更新する。
// urlPrivate は Slack の file.url_private（Bot トークンでの認証が必要）、
// mimetype は Slack が申告する file.mimetype（空ならレスポンスヘッダで代替する）。
func RegisterEventImageFromSlack(ctx context.Context, eventID uint, urlPrivate string, mimetype string) (model.Event, error) {
	event := model.Event{}
	event.ID = eventID
	if err := event.ReadByID(); err != nil {
		return event, fmt.Errorf("failed to read event %d: %w", eventID, err)
	}

	body, headerContentType, err := fetchSlackFile(ctx, urlPrivate)
	if err != nil {
		return event, err
	}
	defer body.Close()

	contentType := mimetype
	if contentType == "" {
		contentType = headerContentType
	}

	// content-type を先に検証しておくことで、不正な形式をストレージに送らずに済む
	if _, err := lib.ExtensionForImageContentType(contentType); err != nil {
		return event, err
	}

	// SDK にシーク可能なリーダーを渡せるよう、上限付きでメモリに読み切る
	data, err := readAllLimited(body, maxEventImageBytes)
	if err != nil {
		return event, err
	}

	previousKey := event.ImageKey

	key, err := lib.EventImages.PutEventImage(ctx, event.ID, contentType, bytes.NewReader(data))
	if err != nil {
		return event, err
	}

	if err := event.UpdateImageKey(key); err != nil {
		return event, fmt.Errorf("failed to update image_key of event %d: %w", eventID, err)
	}

	// 拡張子が変わると新しいキーになるため、古いオブジェクトが残らないよう削除する
	if previousKey != nil && *previousKey != "" && *previousKey != key {
		if err := lib.EventImages.DeleteEventImage(ctx, *previousKey); err != nil {
			log.Printf("failed to delete previous event image (event %d, key %s): %v", eventID, *previousKey, err)
		}
	}
	return event, nil
}

// readAllLimited は limit バイトまで読み込む。超過した場合はエラーを返す
func readAllLimited(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("image is too large (limit %d bytes)", limit)
	}
	return data, nil
}

// fetchSlackFile は url_private を Bot トークン付きで取得する。
// urlPrivate は署名未検証の interaction payload に由来しうるため、
// Slack のファイルホスト以外へBotトークンが送られないよう厳格に検証する。
func fetchSlackFile(ctx context.Context, urlPrivate string) (body io.ReadCloser, contentType string, err error) {
	if urlPrivate == "" {
		return nil, "", errors.New("file url is empty")
	}
	if err := validateSlackFileURL(urlPrivate); err != nil {
		return nil, "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlPrivate, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+slackBotToken)

	client := &http.Client{
		Timeout: lib.SharedHTTPClient.Timeout,
		// リダイレクト先の検証を避けるため、Slackのファイルホストへのリダイレクトも一切追わない
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download slack file: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, "", fmt.Errorf("failed to download slack file: status %d", resp.StatusCode)
	}

	return resp.Body, resp.Header.Get("Content-Type"), nil
}

// validateSlackFileURL は urlPrivate が Slack のファイルホストを指す https URL であることを検証する
func validateSlackFileURL(urlPrivate string) error {
	u, err := url.Parse(urlPrivate)
	if err != nil {
		return fmt.Errorf("invalid file url: %w", err)
	}
	if u.Scheme != "https" {
		return errors.New("file url must use https")
	}
	host := strings.ToLower(u.Hostname())
	if host != "slack.com" && !strings.HasSuffix(host, ".slack.com") {
		return fmt.Errorf("file url host is not a Slack host: %s", host)
	}
	return nil
}
