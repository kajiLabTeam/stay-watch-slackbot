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
    - [GET /api/events/{id}/probability](#get-apieventsidprobability)
      - [リクエスト](#リクエスト-2)
      - [パラメータ](#パラメータ-1)
      - [レスポンス (HTTP 200 OK)](#レスポンス-http-200-ok-1)
      - [使用例](#使用例-2)
  - [Log API](#log-api)
    - [POST /api/logs](#post-apilogs)
      - [リクエスト](#リクエスト-3)
      - [パラメータ](#パラメータ-2)
      - [レスポンス (HTTP 201 Created)](#レスポンス-http-201-created-1)
      - [部分成功時](#部分成功時)
      - [時刻の扱い](#時刻の扱い)
      - [バリデーション](#バリデーション)
      - [使用例](#使用例-3)

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
      "event_id": "0437ac48be2a81",
      "status_id": 1,
      "event_time": "2025-11-24T17:40:26+09:00",
      "room_users": [1001, 1002],
      "participate_users": [1001]
    },
    {
      "event_id": "0437ac48be2a81",
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
| logs[].event_id | string | Yes | イベント識別子（events.code に対応） |
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

1. `event_id` が events テーブルの `code` カラムに存在すること
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
        "event_id": "0437ac48be2a81",
        "status_id": 1,
        "event_time": "2025-11-24T17:40:26+09:00",
        "room_users": [1001, 1002],
        "participate_users": [1001]
      },
      {
        "event_id": "0437ac48be2a81",
        "status_id": 2,
        "event_time": "2025-11-24T17:42:26+09:00",
        "room_users": [],
        "participate_users": []
      }
    ]
  }'
```
