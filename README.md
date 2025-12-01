# Stay Watch Slackbot

研究室メンバーの来訪予測に基づいて、同じ話題に興味のある人が集まる時間帯をSlackで自動通知するボットです。

## 概要

Stay Watch Slackbotは、StayWatch（来訪予測システム）のAPIを利用して、研究室メンバーの来訪確率や時間を予測し、共通の話題（タグ）を持つメンバーが集まりやすい時間帯をSlackのDMで通知します。これにより、自然な交流とコラボレーションを促進します。

![システム構成図](./image.png)

## 主要機能

### 1. ユーザー管理
- SlackユーザーとStayWatch IDの紐付け
- スラッシュコマンドによる簡単なユーザー登録

### 2. タグ（話題）管理
- 「スマブラ」「Android」などの興味のある話題を登録
- 各タグに最低必要人数を設定可能
- モーダルUIによる直感的な登録

### 3. ユーザーとタグの対応付け
- ユーザーが興味のある話題を複数選択可能
- チェックボックスUIで簡単に設定

### 4. 来訪予測・自動通知
- StayWatch APIから各ユーザーの来訪確率と時間を取得
- タグごとに最低人数以上が集まる時間帯を自動計算
- 該当ユーザーにDMで通知
  - 例：「12/2(月) 14:00〜18:00 に `スマブラ` の仲間が集まりそうです」

### 5. インタラクティブな確率照会
- ボットにメンションすると、ユーザー選択UIが表示
- 特定ユーザーの来訪確率をリアルタイムで確認可能

## 技術スタック

- **言語**: Go 1.24.1
- **Webフレームワーク**: Gin v1.10.0
- **データベース**: MySQL 8
- **ORM**: GORM v1.25.12
- **Slack SDK**: slack-go/slack v0.16.0
- **設定管理**: Viper v1.20.1
- **インフラ**: Docker / Docker Compose

## 前提条件

- Docker & Docker Compose
- Slack Workspace（管理者権限）
- StayWatch APIへのアクセス権

## セットアップ

### 1. リポジトリのクローン

```bash
git clone https://github.com/kajiLabTeam/stay-watch-slackbot.git
cd stay-watch-slackbot
```

### 2. 環境変数の設定

`.env`ファイルを編集して、環境変数を設定します：

```bash
# Go API
API_CONTAINER_NAME=staywatch_slackbot_go
API_PORT=8085
GIN_MODE=release

# MySQL
MYSQL_CONTAINER_NAME=staywatch_slackbot_db
MYSQL_ROOT_PASS=your_root_password
MYSQL_USER=your_user
MYSQL_PASS=your_password
MYSQL_DB=app
MYSQL_PORT=3307
```

### 3. Slack設定ファイルの作成

`src/conf/environments/slack.yml`を作成します（テンプレートをコピー）：

```bash
cd src/conf/environments
cp slack_template.yml slack.yml
```

`slack.yml`を編集して、Slack APIの認証情報を設定します：

```yaml
slack:
  signing_secret: "your_signing_secret"
  bot_user_oauth_token: "xoxb-your-bot-token"
```

#### Slack Appの設定

1. [Slack API](https://api.slack.com/apps)でアプリを作成
2. **Bot Token Scopes**に以下を追加：
   - `app_mentions:read`
   - `chat:write`
   - `commands`
   - `im:write`
   - `users:read`
3. **Event Subscriptions**を有効化し、以下を追加：
   - `app_mention`
4. **Slash Commands**を作成：
   - `/add_user` → `https://your-domain.com/slack/command/add_user`
   - `/add_tag` → `https://your-domain.com/slack/command/add_tag`
   - `/add_correspond` → `https://your-domain.com/slack/command/add_correspond`
5. **Interactivity & Shortcuts**を有効化：
   - Request URL: `https://your-domain.com/slack/interaction`

### 4. StayWatch設定ファイルの作成

`src/conf/environments/staywatch.yml`を作成します：

```bash
cp staywatch_template.yml staywatch.yml
```

`staywatch.yml`を編集して、StayWatch APIの情報を設定します：

```yaml
staywatch:
  url: "https://your-staywatch-api.com"
  users: "/api/users"
  probability: "/api/probability"
  time: "/api/time"
```

### 5. MySQL設定ファイルの作成

`src/conf/environments/mysql.yml`を作成します：

```bash
cp mysql_template.yml mysql.yml
```

`mysql.yml`を編集します：

```yaml
mysql:
  user: "your_user"
  password: "your_password"
  protocol: "tcp(db:3306)"
  dbname: "app"
```

### 6. アプリケーションの起動

```bash
# プロジェクトルートで実行
docker compose up -d
```

アプリケーションは `http://localhost:8085` で起動します。

## 使い方

### ユーザーの登録

Slackで以下のコマンドを実行：

```
/add_user <StayWatch User ID>
```

例：
```
/add_user 123
```

### タグ（話題）の登録

```
/add_tag
```

モーダルが開くので、以下を入力：
- **話題**: スマブラ、Android、機械学習など
- **最低人数**: 集まるのに必要な最低人数（例：2）

### タグへの参加

```
/add_correspond
```

モーダルが開くので、興味のあるタグを複数選択できます。

### 来訪確率の確認

ボットをメンション：

```
@StayWatchBot
```

ユーザー選択UIが表示されるので、確認したいユーザーを選択すると、来訪確率が表示されます。

### 自動通知

`GET /notification`エンドポイントを定期的に実行することで、条件に合致したユーザーに自動でDM通知が送信されます。

Google Apps ScriptやCronジョブで定期実行を設定してください：

```bash
curl http://localhost:8085/notification
```

## API エンドポイント

| メソッド | エンドポイント | 説明 |
|---------|--------------|------|
| POST | `/slack/events` | Slackイベントの受信 |
| POST | `/slack/interaction` | Slackインタラクションの処理 |
| POST | `/slack/command/test` | テストコマンド |
| POST | `/slack/command/add_user` | ユーザー登録コマンド |
| POST | `/slack/command/add_tag` | タグ登録コマンド |
| POST | `/slack/command/add_correspond` | ユーザーとタグの対応付けコマンド |
| GET | `/notification` | 条件に合致したユーザーへのDM送信 |

## データベース構造

### Userテーブル

| カラム | 型 | 説明 |
|--------|-----|------|
| ID | uint | 主キー |
| Name | string | ユーザー名 |
| SlackID | string | SlackユーザーID |
| StayWatchID | int64 | StayWatchユーザーID |

### Tagテーブル

| カラム | 型 | 説明 |
|--------|-----|------|
| ID | uint | 主キー |
| Name | string | タグ名（話題） |
| MinNumber | int | 最低必要人数 |

### Correspondテーブル

| カラム | 型 | 説明 |
|--------|-----|------|
| ID | uint | 主キー |
| TagID | uint | タグID（外部キー） |
| UserID | uint | ユーザーID（外部キー） |

UserとTagは多対多の関係を持ちます。

## 開発

### ローカル開発環境

```bash
# 依存関係のインストール
cd src
go mod download

# アプリケーションの実行（ローカル）
go run main.go
```

### ログの確認

```bash
docker compose logs -f api
```

または

```bash
tail -f log/server.log
```

### データベースへの接続

```bash
docker exec -it staywatch_slackbot_db mysql -u kjlb -p app
```

### コンテナの停止

```bash
docker compose down
```

### コンテナの再ビルド

```bash
docker compose up -d --build
```

## プロジェクト構造

```
stay-watch-slackbot/
├── compose.yml              # Docker Compose設定
├── .env                     # 環境変数
├── docker/                  # Dockerファイル
│   ├── app/
│   │   └── dockerfile
│   └── db/
│       └── my.cnf
├── log/                     # ログファイル
│   └── server.log
└── src/                     # ソースコード
    ├── main.go              # エントリーポイント
    ├── go.mod               # Go依存関係管理
    ├── conf/                # 設定管理
    │   ├── config.go
    │   └── environments/    # 環境別設定ファイル
    │       ├── slack.yml
    │       ├── staywatch.yml
    │       └── mysql.yml
    ├── controller/          # HTTPハンドラー
    │   ├── init.go
    │   ├── slack_event.go
    │   ├── slack_interaction.go
    │   ├── slack_comand.go
    │   ├── slack_dm.go
    │   └── GAS_event.go
    ├── service/             # ビジネスロジック
    │   ├── init.go
    │   ├── slack_event.go
    │   ├── user.go
    │   ├── tag.go
    │   ├── correspond.go
    │   └── staywatch.go
    ├── model/               # データモデル
    │   ├── struct.go
    │   ├── user.go
    │   ├── tag.go
    │   └── correspond.go
    ├── lib/                 # ユーティリティ
    │   └── sql.go
    └── router/              # ルーティング
        └── router.go
```

## トラブルシューティング

### ポート競合エラー

別のアプリケーションがポート3307または8085を使用している場合、`.env`ファイルでポート番号を変更してください。

### Slack認証エラー

`slack.yml`のトークンと署名シークレットが正しいか確認してください。

### データベース接続エラー

- `mysql.yml`の設定が`.env`の設定と一致しているか確認
- MySQLコンテナが起動しているか確認: `docker compose ps`

## ライセンス

このプロジェクトは梶研究室チームによって開発されています。

## コントリビューション

プルリクエストを歓迎します。大きな変更の場合は、まずissueを開いて変更内容を議論してください。
