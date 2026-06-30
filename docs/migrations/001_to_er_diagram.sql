-- 活動収集 DB を docs/活動収集db設計.drawio の構成に合わせるためのマイグレーション
--
-- 適用前提: アプリケーションを停止し、必要に応じて DB をバックアップしておくこと。
-- 適用順: 本ファイルを実行 → アプリ起動（AutoMigrate で logs_user_rooms / logs_user_participates が新規作成される）。
--
-- 注意:
-- - corresponds テーブルが存在しない／既に event_users にリネーム済の環境では、RENAME を読み飛ばすか手動で調整すること。
-- - events.type_id を参照している外部キー名は環境によって異なる。実環境では `SHOW CREATE TABLE events;` で確認し、
--   下記の <type_fk_name> プレースホルダを実際の制約名に置き換えてから実行する。
-- - 旧テーブルにデータが残っている場合、本マイグレーションを適用するとそれらは削除される。

START TRANSACTION;

-- 1) corresponds → event_users にリネーム
RENAME TABLE corresponds TO event_users;

-- 2) events から不要カラムを削除し、code を追加
--    code はイベントを一意に定める識別子（UNIQUE）。既存レコードがある環境では
--    DEFAULT '' のまま UNIQUE は付与できないため、データ投入後に
--    `ALTER TABLE events ADD UNIQUE INDEX idx_events_code (code);` を別途適用すること。
ALTER TABLE events
    DROP FOREIGN KEY <type_fk_name>,
    DROP COLUMN type_id,
    ADD COLUMN code VARCHAR(255) NOT NULL DEFAULT '' AFTER name;

-- 3) 多対多および補助テーブルを削除
DROP TABLE IF EXISTS event_tools;
DROP TABLE IF EXISTS tools;
DROP TABLE IF EXISTS types;

COMMIT;

-- 4) アプリ起動後の AutoMigrate により以下が作成される:
--    - logs_user_rooms (id, created_at, updated_at, deleted_at, log_id, user_id)
--    - logs_user_participates (id, created_at, updated_at, deleted_at, log_id, user_id)
--    AutoMigrate は events.code への UNIQUE 制約を後から追加できないため、
--    既存データを移行した後に手動で UNIQUE INDEX を作成する必要がある。
