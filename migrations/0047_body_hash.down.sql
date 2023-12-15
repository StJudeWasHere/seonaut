ALTER TABLE pagereports DROP COLUMN body_hash;
DROP INDEX idx_crawl_id_media_type ON pagereports;
DELETE FROM issue_types WHERE id = 59;