ALTER TABLE organization ADD COLUMN slug TEXT UNIQUE;

CREATE INDEX IF NOT EXISTS Organization_slug ON Organization (slug);
