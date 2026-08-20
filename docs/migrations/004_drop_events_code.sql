-- events.code カラムの削除
--
-- model.Event.Code フィールドの廃止に伴う変更。運用上未使用（外部連携の
-- logs.event_id は events.id を参照しており、code は参照されていなかった）
-- のため削除する。
--
-- GORM の AutoMigrate はカラムの削除を行わないため、本ALTERは手動実行が必要。

ALTER TABLE events
    DROP INDEX idx_events_code,
    DROP COLUMN code;
