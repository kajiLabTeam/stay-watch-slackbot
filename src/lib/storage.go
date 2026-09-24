package lib

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kajiLabTeam/stay-watch-slackbot/config"
)

// EventImageStore は活動画像の保存先を抽象化する。
// 将来ストレージを差し替えても呼び出し側が壊れないよう、必要最小限の2メソッドのみを公開する。
type EventImageStore interface {
	// PutEventImage は画像を保存し、保存先のキーを返す
	PutEventImage(ctx context.Context, eventID uint, contentType string, r io.Reader) (key string, err error)
	// PublicURL はキーに対応する公開URLを返す。未設定・キー未登録の場合は空文字を返す
	PublicURL(key string) string
	// DeleteEventImage はキーに対応するオブジェクトを削除する。キーが空の場合は何もしない
	DeleteEventImage(ctx context.Context, key string) error
}

// EventImages は活動画像ストア。S3系の環境変数が未設定の場合は
// アップロードを拒否しつつ公開URLを空文字で返す no-op 実装になる。
// ストレージ未構築でもボードが落ちないようにするための設計。
var EventImages EventImageStore

// ErrStorageDisabled はストレージが未設定のときに返される
var ErrStorageDisabled = fmt.Errorf("object storage is not configured")

func init() {
	if !config.S3.Enabled() {
		log.Println("object storage is not configured; event images are disabled")
		EventImages = disabledEventImageStore{}
		return
	}
	EventImages = newS3EventImageStore(config.S3)
}

// disabledEventImageStore はストレージ未設定時のフォールバック実装
type disabledEventImageStore struct{}

func (disabledEventImageStore) PutEventImage(_ context.Context, _ uint, _ string, _ io.Reader) (string, error) {
	return "", ErrStorageDisabled
}

func (disabledEventImageStore) PublicURL(_ string) string { return "" }

func (disabledEventImageStore) DeleteEventImage(_ context.Context, _ string) error { return nil }

// s3EventImageStore は S3互換ストレージ（RustFS）への実装
type s3EventImageStore struct {
	client        *s3.Client
	bucket        string
	publicBaseURL string
}

func newS3EventImageStore(cfg config.S3Config) *s3EventImageStore {
	client := s3.New(s3.Options{
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretKey, ""),
		BaseEndpoint: aws.String(strings.TrimRight(cfg.Endpoint, "/")),
		// RustFS は単一ホスト名で運用するため path-style が必須
		UsePathStyle: true,
	})

	return &s3EventImageStore{
		client:        client,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}
}

// PutEventImage は events/{eventID}.{ext} のキーで画像を保存する。
// 拡張子はcontent-typeに依存するため、同一イベントでも形式を変えて
// 再登録すると別キーになる（呼び出し側で旧キーの削除が必要）。
func (s *s3EventImageStore) PutEventImage(ctx context.Context, eventID uint, contentType string, r io.Reader) (string, error) {
	ext, err := ExtensionForImageContentType(contentType)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("events/%d%s", eventID, ext)
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to put object %s: %w", key, err)
	}
	return key, nil
}

// PublicURL はキーに対応する公開URLを返す。
// S3_PUBLIC_BASE_URL が未設定、またはキーが空の場合は空文字を返す。
func (s *s3EventImageStore) PublicURL(key string) string {
	if key == "" || s.publicBaseURL == "" {
		return ""
	}
	return s.publicBaseURL + "/" + strings.TrimLeft(key, "/")
}

// DeleteEventImage はキーに対応するオブジェクトを削除する。キーが空の場合は何もしない
func (s *s3EventImageStore) DeleteEventImage(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("failed to delete object %s: %w", key, err)
	}
	return nil
}

// ExtensionForImageContentType は content-type から拡張子を決める。
// 活動画像として許可するのは png / jpeg のみ。
func ExtensionForImageContentType(contentType string) (string, error) {
	// "image/png; charset=..." のようなパラメータ付きにも対応する
	mediaType := strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))

	switch mediaType {
	case "image/png":
		return ".png", nil
	case "image/jpeg", "image/jpg":
		return ".jpg", nil
	default:
		return "", fmt.Errorf("unsupported image content type: %s", contentType)
	}
}
