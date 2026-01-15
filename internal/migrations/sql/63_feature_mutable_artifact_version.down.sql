ALTER TYPE FEATURE RENAME TO FEATURE_OLD;
CREATE TYPE FEATURE AS ENUM ('licensing', 'pre_post_scripts');

UPDATE Organization SET features = array_remove(features, 'artifact_version_mutable') WHERE 'artifact_version_mutable' = ANY(features);

ALTER TABLE Organization ALTER COLUMN features DROP DEFAULT; -- otherwise the following wouldnt work:
ALTER TABLE Organization
  ALTER COLUMN features TYPE FEATURE[]
    USING (features::text[]::FEATURE[]);
ALTER TABLE Organization
  ALTER COLUMN features SET DEFAULT ARRAY[]::FEATURE[];

DROP TYPE FEATURE_OLD;
