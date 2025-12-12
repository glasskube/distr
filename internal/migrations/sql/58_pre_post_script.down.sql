ALTER TYPE FEATURE RENAME TO FEATURE_OLD;

CREATE TYPE FEATURE AS ENUM ('licensing');

UPDATE Organization SET features = array_remove(features, 'pre_post_scripts') WHERE 'pre_post_scripts' = ANY(features);

ALTER TABLE Organization ALTER COLUMN features DROP DEFAULT; -- otherwise the following wouldnt work:
ALTER TABLE Organization
  ALTER COLUMN features TYPE FEATURE[]
    USING (features::text[]::FEATURE[]);
ALTER TABLE Organization
  ALTER COLUMN features SET DEFAULT ARRAY[]::FEATURE[];

DROP TYPE FEATURE_OLD;

ALTER TABLE Organization
  DROP COLUMN pre_connect_script,
  DROP COLUMN post_connect_script;
