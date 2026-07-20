-- 活動収集 DB を docs/活動収集db設計.drawio の構成に合わせるためのマイグレーション
--
-- 適用前提: アプリケーションを停止し、必要に応じて DB をバックアップしておくこと。
-- 適用順: 本ファイルを上から順に実行 → アプリ起動（AutoMigrate で logs_user_rooms /
--         logs_user_participates が新規作成される）。
--
-- 注意:
-- - corresponds テーブルが存在しない／既に event_users にリネーム済の環境では Step 1 を読み飛ばす。
-- - MySQL の DDL は暗黙コミットされるため、途中でエラーが発生した場合はロールバックできない。
--   各ステップを順番に確認しながら実行すること。

-- ============================================================
-- Step 0: events テーブルの type_id 外部キー制約名を確認する（手動実行）
--
-- SHOW CREATE TABLE events;
--
-- 出力の CONSTRAINT ... FOREIGN KEY (type_id) 行から制約名を確認し、
-- Step 2 の <type_fk_name> を実際の制約名に置き換えてから Step 2 を実行する。
-- ============================================================

-- Step 1: corresponds → event_users にリネーム
RENAME TABLE corresponds TO event_users;

-- Step 2: events の type_id 外部キーを削除する
--         Step 0 で確認した制約名を <type_fk_name> に置き換えてから実行すること
ALTER TABLE events DROP FOREIGN KEY <type_fk_name>;

-- Step 3: 不要カラムを削除し、code カラムを追加する
ALTER TABLE events
    DROP COLUMN type_id,
    ADD COLUMN code VARCHAR(255) NOT NULL DEFAULT '' AFTER name;

-- Step 4: 既存の events 行に一意な code を設定する
--         code が空のまま UNIQUE INDEX を付与すると重複エラーになるため必須
UPDATE events SET code = CONCAT('event_', id) WHERE code = '';

-- Step 5: 多対多および補助テーブルを削除
DROP TABLE IF EXISTS event_tools;
DROP TABLE IF EXISTS tools;
DROP TABLE IF EXISTS types;

-- Step 6: code に UNIQUE INDEX を付与する
--         GORM の AutoMigrate は既存カラムへの UNIQUE 制約追加に対応していないため手動実行が必要
ALTER TABLE events ADD UNIQUE INDEX idx_events_code (code);

-- Step 7: アプリを起動すると AutoMigrate により以下が作成される
--   - logs_user_rooms       (id, created_at, updated_at, deleted_at, log_id, user_id)
--   - logs_user_participates (id, created_at, updated_at, deleted_at, log_id, user_id)
--   - event_users の (event_id, user_id) 複合一意インデックス（idx_event_users_event_user）
