ALTER TABLE employee ADD COLUMN IF NOT EXISTS archive_reason TEXT;
ALTER TABLE employee ADD COLUMN IF NOT EXISTS deleted_by UUID REFERENCES users(id);