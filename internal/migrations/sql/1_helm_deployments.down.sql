ALTER TABLE ApplicationVersion
  DROP COLUMN chart_type,
  DROP COLUMN chart_name,
  DROP COLUMN chart_url,
  DROP COLUMN chart_version,
  DROP COLUMN values_file_data,
  DROP COLUMN template_file_data;

DROP TYPE IF EXISTS HELM_CHART_TYPE CASCADE;
