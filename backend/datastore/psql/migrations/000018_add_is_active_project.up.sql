BEGIN;
ALTER TABLE projects ADD COLUMN is_active boolean NOT NULL DEFAULT true;
CREATE INDEX idx_projects_id_active ON projects (id) WHERE is_active = true;
CREATE INDEX idx_keywords_id_active ON keywords (id) WHERE deleted_at IS NULL;
CREATE INDEX idx_sources_id_active ON sources (id) WHERE deleted_at IS NULL;
COMMIT;