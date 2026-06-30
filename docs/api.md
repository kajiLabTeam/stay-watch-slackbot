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
  - [Type API](#type-api)
    - [GET /api/types](#get-apitypes)
      - [リクエスト](#リクエスト-2)
      - [レスポンス (HTTP 200 OK)](#レスポンス-http-200-ok-1)
      - [使用例](#使用例-2)
    - [POST /api/types](#post-apitypes)
      - [リクエスト](#リクエスト-3)
      - [パラメータ](#パラメータ-1)
      - [レスポンス (HTTP 201 Created)](#レスポンス-http-201-created-1)
      - [使用例](#使用例-3)
  - [Tool API](#tool-api)
    - [GET /api/tools](#get-apitools)
      - [リクエスト](#リクエスト-4)
      - [レスポンス (HTTP 200 OK)](#レスポンス-http-200-ok-2)
      - [使用例](#使用例-4)
    - [POST /api/tools](#post-apitools)
      - [リクエスト](#リクエスト-5)
      - [パラメータ](#パラメータ-2)
      - [レスポンス (HTTP 201 Created)](#レスポンス-http-201-created-2)
      - [使用例](#使用例-5)
  - [Event API](#event-api)
    - [GET /api/events/{id}/probability](#get-apieventsidprobability)
      - [リクエスト](#リクエスト-6)
      - [パラメータ](#パラメータ-3)
      - [レスポンス (HTTP 200 OK)](#レスポンス-http-200-ok-3)
      - [使用例](#使用例-6)
  - [Log API](#log-api)
    - [POST /api/logs](#post-apilogs)
      - [リクエスト](#リクエスト-7)
      - [パラメータ](#パラメータ-4)
      - [レスポンス (HTTP 201 Created)](#レスポンス-http-201-created-3)
      - [部分成功時](#部分成功時)
      - [時刻の扱い](#時刻の扱い)
      - [バリデーション](#バリデーション)
      - [使用例](#使用例-7)

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

## Type API

### GET /api/types

タイプ一覧を取得する。

#### リクエスト

```sh
GET /api/types
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
      "Name": "meeting"
    }
  ]
}
```

#### 使用例

```bash
curl http://localhost:8085/api/types
```

---

### POST /api/types

タイプを一括登録する。

#### リクエスト

```sh
POST /api/types
Content-Type: application/json
```

```json
{
  "names": ["meeting", "study", "work"]
}
```

#### パラメータ

| フィールド | 型 | 必須 | 説明 |
| ----- | ----- | ----- | ----- |
| names | string[] | Yes | タイプ名の配列（1件以上必須） |

#### レスポンス (HTTP 201 Created)

```json
{
  "message": "batch registration completed",
  "data": [...],
  "errors": {}
}
```

#### 使用例

```bash
curl -X POST http://localhost:8085/api/types \
  -H "Content-Type: application/json" \
  -d '{"names": ["meeting", "study", "work"]}'
```

---

## Tool API

### GET /api/tools

ツール一覧を取得する。

#### リクエスト

```sh
GET /api/tools
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
      "Name": "slack"
    }
  ]
}
```

#### 使用例

```bash
curl http://localhost:8085/api/tools
```

---

### POST /api/tools

ツールを一括登録する。

#### リクエスト

```sh
POST /api/tools
Content-Type: application/json
```

```json
{
  "names": ["slack", "discord", "teams"]
}
```

#### パラメータ

| フィールド | 型 | 必須 | 説明 |
| ----- | ----- | ----- | ----- |
| names | string[] | Yes | ツール名の配列（1件以上必須） |

#### レスポンス (HTTP 201 Created)

```json
{
  "message": "batch registration completed",
  "data": [...],
  "errors": {}
}
```

#### 使用例

```bash
curl -X POST http://localhost:8085/api/tools \
  -H "Content-Type: application/json" \
  -d '{"names": ["slack", "discord", "teams"]}'
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
      "event_id": 3,
      "status_id": 1,
      "created_at": "2025-11-24 17:40:26"
    },
    {
      "event_id": 3,
      "status_id": 2,
      "created_at": "2025-11-24 17:42:26"
    }
  ]
}
```

#### パラメータ

| フィールド | 型 | 必須 | 説明 |
| ----- | ----- | ----- | ----- |
| logs | array | Yes | ログエントリの配列（1件以上必須） |
| logs[].event_id | uint | Yes | イベントID（events テーブルに存在する必要あり） |
| logs[].status_id | uint | Yes | ステータスID（statuses テーブルに存在する必要あり） |
| logs[].created_at | string | Yes | 作成日時（JST、形式: `YYYY-MM-DD HH:MM:SS`） |

#### レスポンス (HTTP 201 Created)

```json
{
  "message": "batch registration completed",
  "data": [
    {
      "ID": 1,
      "CreatedAt": "2025-11-24T08:40:26Z",
      "UpdatedAt": "2025-11-24T08:40:26Z",
      "DeletedAt": null,
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

1. `event_id` がeventsテーブルに存在すること
2. `status_id` がstatusesテーブルに存在すること
3. `created_at` が `YYYY-MM-DD HH:MM:SS` 形式であること

#### 使用例

```bash
curl -X POST http://localhost:8085/api/logs \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      {"event_id": 2, "status_id": 1, "created_at": "2025-11-24 17:40:26"},
      {"event_id": 2, "status_id": 2, "created_at": "2025-11-24 17:42:26"}
    ]
  }'
```
