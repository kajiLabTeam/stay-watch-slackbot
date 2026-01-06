# Stay Watch Slackbot

研究室メンバーの来訪予測に基づいて、同じ話題に興味のある人が集まる時間帯をSlackで自動通知するボットです。

## 概要

Stay Watch Slackbotは、StayWatch（来訪予測システム）のAPIを利用して、研究室メンバーの来訪確率や時間を予測し、共通の話題（タグ）を持つメンバーが集まりやすい時間帯をSlackのDMで通知します。これにより、自然な交流とコラボレーションを促進します。

## 主要機能

### 1. ユーザー管理

- SlackユーザーとStayWatch IDの紐付け
- スラッシュコマンドによる簡単なユーザー登録

### 2. イベント管理

- 「スマブラ」「Android」などの興味のある話題を登録
- 各タグに最低必要人数を設定可能
- モーダルUIによる直感的な登録

### 3. ユーザーとイベントの対応付け

- ユーザーが興味のある話題を複数選択可能
- チェックボックスUIで簡単に設定

### 4. イベントの活動履歴管理

- 各イベントごとに過去の活動履歴をデータベースに保存
- 活動履歴を基に、次回の活動予測に活用

### 5. 来訪予測・自動通知

- StayWatch APIから各ユーザーの来訪確率と時間を取得
- イベントの活動履歴から活動の発生しやすい時間帯を分析
- イベントごとに最低人数以上が集まる時間帯を自動計算
- 該当ユーザーにDMで通知
  - 例：17:35〜19:40  スマブラ

### 6. インタラクティブな確率照会

- ボットにメンションすると、ユーザー選択UIが表示
- 特定ユーザーの来訪確率をリアルタイムで確認可能

## 技術スタック

- **言語**: Go 1.24.5
- **Webフレームワーク**: Gin v1.11.0
- **データベース**: MySQL 8.4
- **ORM**: GORM v1.31.1
- **Slack SDK**: slack-go/slack v0.17.3
- **インフラ**:
  - Docker / Docker Compose
  - マルチステージビルド（開発/本番環境分離）
  - 開発環境: air（ホットリロード）
  - 本番環境: Alpine Linux（軽量コンテナ）

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
  api_key: "your_api_key_here"
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

#### 開発環境（air でホットリロード）

```bash
# プロジェクトルートで実行
docker compose up -d
```

開発環境ではairを使用したホットリロードが有効になっており、ソースコードの変更が自動的に反映されます。

#### 本番環境（コンパイル済みバイナリ）

```bash
# 本番環境用のComposeファイルを使用
docker compose -f compose.prod.yml up -d
```

本番環境ではGoバイナリをコンパイルした軽量なAlpineベースのコンテナが起動します。
コンテナクラッシュ時やサーバー再起動時に自動的に再起動されます。

アプリケーションは `http://localhost:8085` で起動します。

## 使い方

### ユーザーの登録

Slackコマンドで自身を登録

``` sh
/add_user
```

### イベントの登録

``` sh
/add_tag
```

モーダルから各種情報を入力

- **イベント名**: スマブラ、カタン、料理会など
- **最低人数**: 集まるのに必要な最低人数（例：2）

### イベントへの参加

``` sh
/add_correspond
```

モーダルから興味のあるイベントを複数選択できます。

### 来訪確率の確認

ボットをメンション：

``` sh
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
| --------- | -------------- | ------ |
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
| --------- | ----- | ------ |
| ID | uint | 主キー（gorm.Model） |
| Name | string | ユーザー名 |
| SlackID | string | SlackユーザーID |
| StayWatchID | int64 | StayWatchユーザーID |
| Corresponds | []Correspond | ユーザーが参加するイベント |

### Eventテーブル

| カラム | 型 | 説明 |
| --------- | ----- | ------ |
| ID | uint | 主キー（gorm.Model） |
| Name | string | イベント名（スマブラ、カタンなど） |
| MinNumber | int | 最低必要人数（デフォルト: 2） |
| TypeID | uint | イベントタイプID（外部キー） |
| Type | Type | イベントタイプ |
| Tools | []Tool | 使用するツール（多対多） |
| Corresponds | []Correspond | イベント参加者 |

### Typeテーブル

| カラム | 型 | 説明 |
| --------- | ----- | ------ |
| ID | uint | 主キー（gorm.Model） |
| Name | string | イベントタイプ名 |
| Events | []Event | このタイプに属するイベント |

### Toolテーブル

| カラム | 型 | 説明 |
| --------- | ----- | ------ |
| ID | uint | 主キー（gorm.Model） |
| Name | string | ツール名 |
| Events | []Event | このツールを使用するイベント（多対多） |

### Correspondテーブル

| カラム | 型 | 説明 |
| --------- | ----- | ------ |
| ID | uint | 主キー（gorm.Model） |
| EventID | uint | イベントID（外部キー） |
| Event | Event | イベント |
| UserID | uint | ユーザーID（外部キー） |
| User | User | ユーザー |

### Statusテーブル

| カラム | 型 | 説明 |
| --------- | ----- | ------ |
| ID | uint | 主キー（gorm.Model） |
| Name | string | ステータス名（start, end, pose） |
| Logs | []Log | このステータスに関連するログ |

### Logテーブル

| カラム | 型 | 説明 |
| --------- | ----- | ------ |
| ID | uint | 主キー（gorm.Model） |
| EventID | uint | イベントID（外部キー） |
| Event | Event | イベント |
| StatusID | uint | ステータスID（外部キー） |
| Status | Status | ステータス |

### UserDetailテーブル

| カラム | 型 | 説明 |
| --------- | ----- | ------ |
| User | User | ユーザー情報 |
| VisitProbability | float64 | 来訪確率 |
| VisitTime | string | 来訪時刻 |
| DepartureTime | string | 退出時刻 |

### リレーション

- **User ↔ Event**: 多対多（Correspondテーブルで関連付け）
- **Event ↔ Type**: 多対1（EventはTypeに属する）
- **Event ↔ Tool**: 多対多（event_toolsテーブルで関連付け）
- **Event ↔ Log**: 1対多（Eventは複数のLogを持つ）
- **Status ↔ Log**: 1対多（Statusは複数のLogを持つ）

## 開発

### Docker環境について

このプロジェクトでは、開発環境と本番環境でDocker構成を分離しています。

#### 開発環境の特徴

- **ホットリロード**: airによる自動リロード機能
- **ソースマウント**: ローカルのソースコードをコンテナにマウント
- **高速な開発サイクル**: コードの変更がすぐに反映

#### 本番環境の特徴

- **軽量イメージ**: Alpine Linuxベースで最小限のサイズ
- **コンパイル済みバイナリ**: 最適化されたGoバイナリを実行
- **自動再起動**: `restart: unless-stopped`により自動復旧
- **データ永続化**: MySQLデータはDockerボリュームで永続化

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
# 開発環境
docker compose logs -f api

# 本番環境
docker compose -f compose.prod.yml logs -f api
```

または

```bash
tail -f log/server.log
```

### データベースへの接続

```bash
docker exec -it staywatch_slackbot_db mysql -u kjlb -p app
```

### データベースの永続化

MySQLのデータは名前付きボリューム `mysql_data` に永続化されます。
コンテナを削除しても、データは保持されます。

```bash
# ボリュームの確認
docker volume ls | grep mysql_data

# ボリュームの削除（データも削除されるので注意）
docker volume rm stay-watch-slackbot_mysql_data
```

### コンテナの停止

```bash
# 開発環境
docker compose down

# 本番環境
docker compose -f compose.prod.yml down
```

### コンテナの再ビルド

```bash
# 開発環境
docker compose up -d --build

# 本番環境
docker compose -f compose.prod.yml up -d --build
```

## プロジェクト構造

``` sh
stay-watch-slackbot/
├── compose.yml              # Docker Compose設定（開発環境）
├── compose.prod.yml         # Docker Compose設定（本番環境）
├── .env                     # 環境変数
├── docker/                  # Dockerファイル
│   ├── app/
│   │   └── dockerfile       # マルチステージビルド対応
│   └── db/
│       └── my.cnf           # MySQL設定
├── log/                     # ログファイル
│   └── server.log
└── src/                     # ソースコード
    ├── main.go              # エントリーポイント
    ├── go.mod               # Go依存関係管理
    ├── go.sum
    ├── .air.toml            # air設定（開発環境用）
    ├── conf/                # 設定管理
    │   ├── config.go
    │   └── environments/    # 環境別設定ファイル
    │       ├── slack.yml
    │       ├── staywatch.yml
    │       └── mysql.yml
    ├── controller/          # HTTPハンドラー
    │   ├── init.go
    │   ├── helper.go
    │   ├── slack_event.go
    │   ├── slack_interaction.go
    │   ├── slack_command.go
    │   ├── slack_dm.go
    │   └── GAS_event.go
    ├── service/             # ビジネスロジック
    │   ├── init.go
    │   ├── slack_event.go
    │   ├── user.go
    │   ├── event.go
    │   ├── correspond.go
    │   ├── staywatch.go
    │   ├── activity.go      # 活動履歴管理
    │   ├── notification.go  # 通知処理
    │   └── occupancy.go     # 在室状況管理
    ├── model/               # データモデル
    │   ├── struct.go        # 全モデル定義
    │   ├── user.go
    │   ├── event.go
    │   ├── correspond.go
    │   ├── log.go
    │   ├── status.go
    │   ├── tool.go
    │   └── types.go
    ├── prediction/          # 予測アルゴリズム
    │   ├── clustering.go    # クラスタリング
    │   └── probability.go   # 確率計算
    ├── lib/                 # ユーティリティ
    │   ├── sql.go
    │   ├── http_client.go
    │   ├── staywatch_client.go
    │   ├── math_utils.go
    │   └── time_utils.go
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

### コンテナが自動再起動しない（本番環境）

本番環境で `restart: unless-stopped` が設定されているか確認してください。
手動で `docker compose stop` した場合は再起動しません。

### データベースのデータが消える

Docker Composeの設定でボリューム定義が正しいか確認してください。
`docker volume ls` でボリュームが作成されているか確認できます。

## ライセンス

このプロジェクトは梶研究室チームによって開発されています。

## コントリビューション

プルリクエストを歓迎します。大きな変更の場合は、まずissueを開いて変更内容を議論してください。
