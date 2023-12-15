ALTER TABLE pagereports DROP COLUMN body_hash;
DELETE FROM issue_types WHERE id = 59;