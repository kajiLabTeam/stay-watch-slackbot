# DB スキーマ設計

`docs/活動収集db設計.drawio` の ER 図に基づくスキーマ構成。

## エンティティ一覧

| テーブル | 用途 |
| --- | --- |
| `users` | システム利用ユーザー（Slack/StayWatch と紐付け） |
| `events` | 活動イベント（スマブラ、人生ゲーム など） |
| `statuses` | 活動ステータス（start / end / pose） |
| `logs` | 活動ログ（イベントの開始・終了等の記録） |
| `event_users` | Event ↔ User 中間テーブル（イベント担当者） |
| `logs_user_rooms` | Log ↔ User 中間テーブル（ログ発生時に在室していたユーザー） |
| `logs_user_participates` | Log ↔ User 中間テーブル（ログ対象イベントに参加したユーザー） |

すべてのテーブルは GORM の `gorm.Model`（`id`, `created_at`, `updated_at`, `deleted_at`）を含む。

---

## テーブル定義

### users

| カラム | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| `id` | uint | PK | 内部ユーザー ID |
| `created_at` | datetime | | |
| `updated_at` | datetime | | |
| `deleted_at` | datetime | index, nullable | 論理削除 |
| `name` | varchar(255) | | 表示名 |
| `slack_id` | varchar(255) | | Slack ユーザー ID |
| `stay_watch_id` | bigint | | StayWatch システム上の ID |

**関連:**
- `event_users` を介して `events` と多対多
- `logs_user_rooms` を介して `logs` と多対多（在室）
- `logs_user_participates` を介して `logs` と多対多（参加）

---

### events

| カラム | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| `id` | uint | PK | 内部イベント ID |
| `created_at` | datetime | | |
| `updated_at` | datetime | | |
| `deleted_at` | datetime | index, nullable | |
| `code` | varchar(255) | unique, not null | イベントを一意に定める識別子（例: `1`, `2`, `0437ac48be2a81`） |
| `name` | varchar(255) | unique, not null | イベント名（例: スマブラ、人生ゲーム） |
| `min_number` | int | default 2 | 活動成立に必要な最低人数 |

**関連:**
- `event_users` を介して `users` と多対多
- `logs` と一対多（`logs.event_id`）

---

### statuses

| カラム | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| `id` | uint | PK | |
| `created_at` | datetime | | |
| `updated_at` | datetime | | |
| `deleted_at` | datetime | index, nullable | |
| `name` | varchar(255) | unique, not null | `start` / `end` / `pose` |

**関連:**
- `logs` と一対多（`logs.status_id`）

---

### logs

| カラム | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| `id` | uint | PK | |
| `created_at` | datetime | | |
| `updated_at` | datetime | | |
| `deleted_at` | datetime | index, nullable | |
| `event_time` | datetime | | ログイベント発生時刻 |
| `event_id` | uint | FK → `events.id`, ON UPDATE CASCADE / ON DELETE CASCADE | |
| `status_id` | uint | FK → `statuses.id`, ON UPDATE CASCADE / ON DELETE CASCADE | |

**関連:**
- `events` と多対一
- `statuses` と多対一
- `logs_user_rooms` を介して `users` と多対多（在室ユーザー = `room_users`）
- `logs_user_participates` を介して `users` と多対多（参加ユーザー = `participate_users`）

---

### event_users

Event と User の中間テーブル。「誰がどのイベントを担当しているか」を表す。

| カラム | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| `id` | uint | PK | |
| `created_at` | datetime | | |
| `updated_at` | datetime | | |
| `deleted_at` | datetime | index, nullable | |
| `event_id` | uint | FK → `events.id`, ON UPDATE CASCADE / ON DELETE CASCADE | |
| `user_id` | uint | FK → `users.id`, ON UPDATE CASCADE / ON DELETE CASCADE | |

---

### logs_user_rooms

Log と User の中間テーブル。あるログ発生時点で在室していたユーザーを表す。

| カラム | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| `id` | uint | PK | |
| `created_at` | datetime | | |
| `updated_at` | datetime | | |
| `deleted_at` | datetime | index, nullable | |
| `log_id` | uint | FK → `logs.id`, ON UPDATE CASCADE / ON DELETE CASCADE | |
| `user_id` | uint | FK → `users.id`, ON UPDATE CASCADE / ON DELETE CASCADE | |

---

### logs_user_participates

Log と User の中間テーブル。そのイベント（ログ）に参加したユーザーを表す。

| カラム | 型 | 制約 | 説明 |
| --- | --- | --- | --- |
| `id` | uint | PK | |
| `created_at` | datetime | | |
| `updated_at` | datetime | | |
| `deleted_at` | datetime | index, nullable | |
| `log_id` | uint | FK → `logs.id`, ON UPDATE CASCADE / ON DELETE CASCADE | |
| `user_id` | uint | FK → `users.id`, ON UPDATE CASCADE / ON DELETE CASCADE | |

---

## ER 概略

```
[ users ] 1 ─── N [ event_users ] N ─── 1 [ events ] 1 ─── N [ logs ] N ─── 1 [ statuses ]
    │                                                            │
    │ 1                                                        N │
    ├──────── N [ logs_user_rooms ] N ──────────────────────────┤
    │                                                            │
    └──────── N [ logs_user_participates ] N ───────────────────┘
```

- Event 1 : N Logs
- Status 1 : N Logs
- Event N : M User （`event_users`）
- Log N : M User （`logs_user_rooms`：在室ユーザー）
- Log N : M User （`logs_user_participates`：参加ユーザー）

---

## 現行スキーマからの差分

### 削除されるテーブル / カラム

- `types` テーブル（および `events.type_id`）
- `tools` テーブル
- `event_tools`（Event ↔ Tool の多対多）
- （`events.min_number` は存続）

### 追加されるテーブル / カラム

- `events.code`（イベントを一意に定める文字列 ID。`uniqueIndex; not null`）
- `logs_user_rooms`
- `logs_user_participates`

### リネーム

- `corresponds` → `event_users`（役割は同一：Event ↔ User の中間テーブル）
