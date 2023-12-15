ALTER TABLE pagereports ADD COLUMN body_hash CHAR(64) NOT NULL;
CREATE INDEX idx_body_hash ON pagereports (body_hash);
CREATE INDEX idx_crawl_id_media_type ON pagereports (crawl_id, media_type);

INSERT INTO issue_types (id, type, priority) VALUES(59, "ERROR_DUPLICATED_CONTENT", 2);