# REST API リファレンス

stay-watch-slackbotのREST APIドキュメント。

## 目次

- [REST API リファレンス](#rest-api-リファレンス)
  - [目次](#目次)
  - [Status API](#status-api)
    - [GET /api/statuses](#get-apistatuses)
      - [リクエスト](#リクエスト)
      - [レスポンス (HTTP 200 OK)](#レスポンス-http-200-ok)
      - [使用例](#使用例)
    - [POST /api/statuses](#post-apistatuses)
      - [リクエスト](#リクエスト-1)
      - [パラメータ](#パラメータ)
      - [レスポンス (HTTP 201 Created)](#レスポンス-http-201-created)
      - [使用例](#使用例-1)
  - [Event API](#event-api)
    - [GET /api/events](#get-apievents)
      - [リクエスト](#リクエスト-2)
      - [パラメータ](#パラメータ-1)
      - [レスポンス (HTTP 200 OK)](#レスポンス-http-200-ok-1)
      - [使用例](#使用例-2)
    - [GET /api/events/{id}/probability](#get-apieventsidprobability)
      - [リクエスト](#リクエスト-3)
      - [パラメータ](#パラメータ-2)
      - [レスポンス (HTTP 200 OK)](#レスポンス-http-200-ok-2)
      - [使用例](#使用例-3)
  - [Log API](#log-api)
    - [POST /api/logs](#post-apilogs)
      - [リクエスト](#リクエスト-4)
      - [パラメータ](#パラメータ-3)
      - [レスポンス (HTTP 201 Created)](#レスポンス-http-201-created-1)
      - [部分成功時](#部分成功時)
      - [時刻の扱い](#時刻の扱い)
      - [バリデーション](#バリデーション)
      - [使用例](#使用例-4)
  - [Board API](#board-api)
    - [GET /api/board](#get-apiboard)
      - [リクエスト](#リクエスト-5)
      - [レスポンス (HTTP 200 OK)](#レスポンス-http-200-ok-3)
      - [使用例](#使用例-5)
  - [Slack Command API](#slack-command-api)
    - [POST /slack/command/add_event_image](#post-slackcommandadd_event_image)
      - [リクエスト](#リクエスト-6)
      - [レスポンス](#レスポンス)

---

## Status API

### GET /api/statuses

ステータス一覧を取得する。

#### リクエスト

```sh
GET /api/statuses
```

#### レスポンス (HTTP 200 OK)

```json
{
  "data": [
    {
      "ID": 1,
      "CreatedAt": "2025-01-01T00:00:00Z",
      "UpdatedAt": "2025-01-01T00:00:00Z",
      "DeletedAt": null,
      "Name": "start"
    },
    {
      "ID": 2,
      "CreatedAt": "2025-01-01T00:00:00Z",
      "UpdatedAt": "2025-01-01T00:00:00Z",
      "DeletedAt": null,
      "Name": "end"
    }
  ]
}
```

#### 使用例

```bash
curl http://localhost:8085/api/statuses
```

---

### POST /api/statuses

ステータスを一括登録する。

#### リクエスト

```sh
POST /api/statuses
Content-Type: application/json
```

```json
{
  "names": ["start", "end", "pause"]
}
```

#### パラメータ

| フィールド | 型 | 必須 | 説明 |
| ----- | ----- | ----- | ----- |
| names | string[] | Yes | ステータス名の配列（1件以上必須） |

#### レスポンス (HTTP 201 Created)

```json
{
  "message": "batch registration completed",
  "data": [
    {
      "ID": 1,
      "CreatedAt": "2025-01-01T00:00:00Z",
      "UpdatedAt": "2025-01-01T00:00:00Z",
      "DeletedAt": null,
      "Name": "start"
    }
  ],
  "errors": {
    "end": "status already exists"
  }
}
```

#### 使用例

```bash
curl -X POST http://localhost:8085/api/statuses \
  -H "Content-Type: application/json" \
  -d '{"names": ["start", "end", "pause"]}'
```

---

## Event API

### GET /api/events

イベント一覧を取得する。`id` または `game_type` を指定すると絞り込みができる。

#### リクエスト

```sh
GET /api/events
GET /api/events?id=5
GET /api/events?game_type=digital
```

#### パラメータ

| パラメータ | 型 | 必須 | 説明 |
| ----- | ----- | ----- | ----- |
| id | uint | No | 指定したIDのイベント1件のみを取得する |
| game_type | string | No | `digital` または `analog` を指定し、その分類のイベントのみに絞り込む |

`id` と `game_type` を同時に指定した場合は `id` が優先される。いずれも指定しない場合は全件を返す。

#### レスポンス (HTTP 200 OK)

全件・`game_type` 指定時（配列）:

```json
{
  "data": [
    {
      "ID": 2,
      "CreatedAt": "2026-01-20T03:52:00.051Z",
      "UpdatedAt": "2026-01-20T03:52:00.051Z",
      "DeletedAt": null,
      "Name": "大乱闘スマッシュブラザーズ",
      "MinNumber": 3,
      "GameTypeID": 1,
      "GameType": {
        "ID": 1,
        "Name": "digital"
      }
    }
  ]
}
```

`id` 指定時（単一オブジェクト）:

```json
{
  "data": {
    "ID": 5,
    "Name": "カタン(スタンダート)",
    "MinNumber": 3,
    "GameTypeID": 2,
    "GameType": {
      "ID": 2,
      "Name": "analog"
    }
  }
}
```

`id` に該当するイベントが存在しない場合は `404 Not Found`、`game_type` に `digital`/`analog` 以外を指定した場合は `400 Bad Request` を返す。

#### 使用例

```bash
curl http://localhost:8085/api/events
curl "http://localhost:8085/api/events?id=5"
curl "http://localhost:8085/api/events?game_type=analog"
```

---

### GET /api/events/{id}/probability

指定したイベントの発生確率を取得する。

#### リクエスト

```sh
GET /api/events/{id}/probability?weekday=0&time=12:00
```

#### パラメータ

| パラメータ | 型 | 必須 | 説明 |
| ----- | ----- | ----- | ----- |
| id | uint | Yes | イベントID（パスパラメータ） |
| weekday | int | Yes | 曜日（0=月曜日, 6=日曜日） |
| time | string | No | 時刻（JST、形式: `HH:MM`、デフォルト: 現在時刻） |

#### レスポンス (HTTP 200 OK)

```json
{
  "event_id": 1,
  "weekday": 0,
  "time": "12:00",
  "probability": 0.75
}
```

#### 使用例

```bash
curl "http://localhost:8085/api/events/1/probability?weekday=0&time=12:00"
```

---

## Log API

### POST /api/logs

外部システムからログを一括登録する。

#### リクエスト

```sh
POST /api/logs
Content-Type: application/json
```

```json
{
  "logs": [
    {
      "event_id": "3",
      "status_id": 1,
      "event_time": "2025-11-24T17:40:26+09:00",
      "room_users": [1001, 1002],
      "participate_users": [1001]
    },
    {
      "event_id": "3",
      "status_id": 2,
      "event_time": "2025-11-24T17:42:26+09:00",
      "room_users": [],
      "participate_users": []
    }
  ]
}
```

#### パラメータ

| フィールド | 型 | 必須 | 説明 |
| ----- | ----- | ----- | ----- |
| logs | array | Yes | ログエントリの配列（1件以上必須） |
| logs[].event_id | string | Yes | イベント識別子（events.id に対応） |
| logs[].status_id | uint | Yes | ステータスID（statuses テーブルに存在する必要あり） |
| logs[].event_time | string | Yes | イベント発生日時（JST、RFC3339形式: `2006-01-02T15:04:05+09:00`） |
| logs[].room_users | int64[] | No | 在室メンバの stay_watch_id の配列（省略可） |
| logs[].participate_users | int64[] | No | 参加メンバの stay_watch_id の配列（省略可） |

#### レスポンス (HTTP 201 Created)

```json
{
  "message": "batch registration completed",
  "data": [
    {
      "ID": 1,
      "CreatedAt": "2025-11-24T17:40:26+09:00",
      "UpdatedAt": "2025-11-24T17:40:26+09:00",
      "DeletedAt": null,
      "EventTime": "2025-11-24T17:40:26+09:00",
      "EventID": 3,
      "Event": {},
      "StatusID": 1,
      "Status": {}
    }
  ],
  "errors": {}
}
```

#### 部分成功時

一部のログが登録に失敗した場合でも、成功したログは登録され、エラーは `errors` フィールドに返される。

```json
{
  "message": "batch registration completed",
  "data": [...],
  "errors": {
    "1": "event_id 999 not found",
    "3": "status_id 100 not found"
  }
}
```

`errors` のキーは入力配列のインデックス（0始まり）。

#### 時刻の扱い

| 処理 | タイムゾーン |
| ----- | ----- |
| 入力 | JST（日本標準時） |
| 保存 | JST |
| 出力 | JST |

システム全体でJSTに統一しています。

例:

- 入力: `2025-11-24T17:40:26+09:00` (JST)
- 保存/出力: `2025-11-24T17:40:26+09:00` (JST)

#### バリデーション

1. `event_id` が events テーブルの `id` カラムに存在すること
2. `status_id` が statuses テーブルに存在すること
3. `event_time` が RFC3339 形式（`+09:00` など UTC オフセット付き）であること
4. `room_users`・`participate_users` の各 stay_watch_id が users テーブルに存在すること

#### 使用例

```bash
curl -X POST http://localhost:8085/api/logs \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      {
        "event_id": "3",
        "status_id": 1,
        "event_time": "2025-11-24T17:40:26+09:00",
        "room_users": [1001, 1002],
        "participate_users": [1001]
      },
      {
        "event_id": "3",
        "status_id": 2,
        "event_time": "2025-11-24T17:42:26+09:00",
        "room_users": [],
        "participate_users": []
      }
    ]
  }'
```

---

## Board API

### GET /api/board

共有モニター(moment-board)用の集約表示データを取得する。

- `hours` は現在時刻を先頭に最大4時間ぶんの在室予想タイムライン。11時より前は11時始まりに丸められ、19時を超える列は出さないため、夜間は列が減っていき最終的に空配列になる。
- `hours[].activities` はその時間帯ごとに成立しそうな活動一覧。「その時間帯に在室していそうなメンバー」が最低人数（`minNumber`）以上そろい、かつその時間帯のGMM活動確率が閾値（`BOARD_ACTIVITY_PROBABILITY_THRESHOLD`、デフォルト `0.3`）以上の活動のみを返す。
- `presence.members` は現在の在室者。StayWatch の在室API（`STAYWATCH_PRESENCE_PATH`、デフォルト `/api/v1/stayers`）から取得し、`stay_watch_id` が一致するユーザーの `name` と `avatarUrl`（`icon_url`）を返す。DB未登録の在室者は StayWatch 上の名前と空の `avatarUrl` で返す。StayWatch の取得に失敗した場合は空配列になる。
- `imageUrl` は画像未登録、またはオブジェクトストレージ未設定の場合に `null` になる。

#### リクエスト

```sh
GET /api/board
```

#### レスポンス (HTTP 200 OK)

```json
{
  "currentTime": "15:04",
  "presence": {
    "members": [
      { "name": "hanada", "avatarUrl": "https://example.com/avatar.png" }
    ]
  },
  "hours": [
    {
      "hour": 15,
      "people": [
        { "name": "enami", "avatarUrl": "https://example.com/avatar.png" }
      ],
      "activities": [
        {
          "id": 5,
          "name": "カタン(スタンダート)",
          "imageUrl": "https://storage.example.com/daycast/events/5.png",
          "minNumber": 3,
          "members": [
            { "name": "hanada", "avatarUrl": "https://example.com/avatar.png" }
          ]
        }
      ]
    },
    { "hour": 16, "people": [], "activities": [] },
    { "hour": 17, "people": [], "activities": [] },
    { "hour": 18, "people": [], "activities": [] }
  ]
}
```

#### 使用例

```bash
curl http://localhost:8085/api/board
```

---

## Slack Command API

### POST /slack/command/add_event_image

Slack のスラッシュコマンド `/add_event_image` を受け取り、活動画像の登録モーダルを開く。

モーダルは対象イベントの `static_select` と、png / jpg / jpeg を1ファイル受け付ける `file_input` で構成される。
送信されると `POST /slack/interaction`（`view_submission`、`callback_id: register_event_image`）に届き、
Slack の `url_private` を Bot トークンで取得してオブジェクトストレージへ保存し、`events.image_key` を更新する。

Slack App 側に **`files:read` スコープ** と スラッシュコマンド `/add_event_image` の登録が必要。

#### リクエスト

Slack からの `application/x-www-form-urlencoded` なスラッシュコマンドペイロード。

#### レスポンス

モーダルを開いた旨のエフェメラルメッセージを返す。登録結果はモーダル送信後に `response_url` へ投稿される。
