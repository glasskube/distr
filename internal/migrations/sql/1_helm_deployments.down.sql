ALTER TABLE ApplicationVersion
  DROP COLUMN chart_type HELM_CHART_TYPE,
  DROP COLUMN chart_name TEXT,
  DROP COLUMN chart_url TEXT,
  DROP COLUMN chart_version TEXT,
  DROP COLUMN values_file_data,
  DROP COLUMN template_file_data;

DROP TYPE IF EXISTS HELM_CHART_TYPE CASCADE;
